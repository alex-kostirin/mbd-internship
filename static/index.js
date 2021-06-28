let map = null
let visibleMetric = 'uniq_users'
let infoWindow = null
let cells = new Map()

function initMap () {
  map = new google.maps.Map(document.getElementById('map'), {
    center: { lat: 24.886, lng: -70.268 },
    mapTypeId: 'terrain',
    zoom: 16,
  })

  map.controls[google.maps.ControlPosition.TOP_CENTER].push(createCustomControl())

  switchToCurrentPosition(map)

  map.addListener('idle', () => {
    drawS2Cells(map).then()
  })
}

function switchToCurrentPosition (map) {
  if (navigator.geolocation) {
    navigator.geolocation.getCurrentPosition(
      (position) => {
        const pos = {
          lat: position.coords.latitude,
          lng: position.coords.longitude,
        }

        map.setCenter(pos)
      },
      () => {
      }
    )
  }
}

function createCustomControl () {
  const template = document.createElement('template')
  template.innerHTML = `
    <div class="radio-control">
      <input type="radio" id="choice-uniq-users"
               name="visibleMetric" value="uniq_users" checked onclick="handleRadioClick(this)">
      <label class="radio-item" for="choice-uniq-users">Uniq users</label>
    
      <input type="radio" id="choice-signal-avg"
             name="visibleMetric" value="signal_avg" onclick="handleRadioClick(this)">
      <label class="radio-item" for="choice-signal-avg">Signal average</label>
    </div>
  `
  return template.content.children[0]
}

function handleRadioClick (e) {
  if (visibleMetric !== e.value) {
    visibleMetric = e.value
    drawS2Cells().then()
  }
}

async function drawS2Cells () {

  if (infoWindow) {
    infoWindow.close()
  }

  const data = await fetchData(map)
  const newCells = clearNotVisibleCells(data)
  const colorScale = getColorScale(data)

  for (const element of data) {
    const key = element['s2_id']
    const metric = element[visibleMetric]
    if (newCells.has(key)) {
      const polygon = newCells.get(key)
      polygon['uniq_users'] = element['uniq_users']
      polygon['signal_avg'] = element['signal_avg']
      polygon.setOptions({
        strokeColor: colorScale(metric),
        fillColor: colorScale(metric),
      })
      continue
    }
    const paths = []
    for (const coordinate of element['s2_coordinates']) {
      paths.push({ lat: coordinate.lat, lng: coordinate.lon })
    }
    const polygon = new google.maps.Polygon({
      paths: paths,
      strokeColor: colorScale(metric),
      strokeOpacity: 0.8,
      strokeWeight: 2,
      fillColor: colorScale(metric),
      fillOpacity: 0.35,
    })
    polygon['s2_id'] = element['s2_id']
    polygon['uniq_users'] = element['uniq_users']
    polygon['signal_avg'] = element['signal_avg']
    newCells.set(key, polygon)
    polygon.setMap(map)
    polygon.addListener('click', handleCellClick)
  }
  cells = newCells
}

async function fetchData (map) {
  const mapBounds = map.getBounds()
  const request = {
    area: [
      {
        lat: mapBounds.getSouthWest().lat(),
        lon: mapBounds.getSouthWest().lng()
      },
      {
        lat: mapBounds.getNorthEast().lat(),
        lon: mapBounds.getNorthEast().lng()
      }
    ]
  }
  const response = await fetch('http://127.0.0.1:8080/data', {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(request)
  })
  return await response.json()
}

function clearNotVisibleCells (data) {
  const newCells = new Map()
  for (const element of data) {
    const key = element['s2_id']

    if (cells.has(key)) {
      newCells.set(key, cells.get(key))
    }
  }
  for (const [key, value] of cells) {
    if (!newCells.has(key)) {
      value.setMap(null)
    }
  }
  return newCells
}

function getColorScale (data) {
  return d3.scaleSequential()
    .interpolator(d3.interpolateWarm)
    .domain([d3.min(data.map((e) => e[visibleMetric])), d3.max(data.map((e) => e[visibleMetric]))])
}

function handleCellClick (e) {
  if (infoWindow === null) {
    infoWindow = new google.maps.InfoWindow({})
  }
  infoWindow.close()
  let metricValue = this[visibleMetric]
  if (metricValue % 1 !== 0) {
    metricValue = metricValue.toFixed(2)
  }
  infoWindow.setContent(`
    <b>${this['s2_id'].toString(16)}</b>
    <br/>
    <div class="center-text">${metricValue}</div>
`)
  infoWindow.setPosition(e.latLng)
  infoWindow.open(this.map)
}

