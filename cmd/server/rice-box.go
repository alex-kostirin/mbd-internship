// Code generated by rice embed-go; DO NOT EDIT.
package main

import (
	"time"

	"github.com/GeertJohan/go.rice/embedded"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "index.html",
		FileModTime: time.Unix(1623911453, 0),

		Content: string("<!DOCTYPE html>\n<html>\n<head>\n    <title>Map</title>\n    <script src=\"https://polyfill.io/v3/polyfill.min.js?features=default\"></script>\n    <script src=\"https://d3js.org/d3.v4.js\"></script>\n    <link rel=\"stylesheet\" type=\"text/css\" href=\"./style.css\"/>\n    <script src=\"./index.js\"></script>\n</head>\n<body>\n<div id=\"map\"></div>\n<script\n        src=\"https://maps.googleapis.com/maps/api/js?key=AIzaSyB41DRUbKWJHPxaFjMAwdrzWzbVKartNGg&callback=initMap&libraries=&v=weekly\"\n        async>\n</script>\n</body>\n</html>\n"),
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    "index.js",
		FileModTime: time.Unix(1624876502, 0),

		Content: string("let map = null\nlet visibleMetric = 'uniq_users'\nlet infoWindow = null\nlet cells = new Map()\n\nfunction initMap () {\n  map = new google.maps.Map(document.getElementById('map'), {\n    center: { lat: 24.886, lng: -70.268 },\n    mapTypeId: 'terrain',\n    zoom: 16,\n  })\n\n  map.controls[google.maps.ControlPosition.TOP_CENTER].push(createCustomControl())\n\n  switchToCurrentPosition(map)\n\n  map.addListener('idle', () => {\n    drawS2Cells(map).then()\n  })\n}\n\nfunction switchToCurrentPosition (map) {\n  if (navigator.geolocation) {\n    navigator.geolocation.getCurrentPosition(\n      (position) => {\n        const pos = {\n          lat: position.coords.latitude,\n          lng: position.coords.longitude,\n        }\n\n        map.setCenter(pos)\n      },\n      () => {\n      }\n    )\n  }\n}\n\nfunction createCustomControl () {\n  const template = document.createElement('template')\n  template.innerHTML = `\n    <div class=\"radio-control\">\n      <input type=\"radio\" id=\"choice-uniq-users\"\n               name=\"visibleMetric\" value=\"uniq_users\" checked onclick=\"handleRadioClick(this)\">\n      <label class=\"radio-item\" for=\"choice-uniq-users\">Uniq users</label>\n    \n      <input type=\"radio\" id=\"choice-signal-avg\"\n             name=\"visibleMetric\" value=\"signal_avg\" onclick=\"handleRadioClick(this)\">\n      <label class=\"radio-item\" for=\"choice-signal-avg\">Signal average</label>\n    </div>\n  `\n  return template.content.children[0]\n}\n\nfunction handleRadioClick (e) {\n  if (visibleMetric !== e.value) {\n    visibleMetric = e.value\n    drawS2Cells().then()\n  }\n}\n\nasync function drawS2Cells () {\n\n  if (infoWindow) {\n    infoWindow.close()\n  }\n\n  const data = await fetchData(map)\n  const newCells = clearNotVisibleCells(data)\n  const colorScale = getColorScale(data)\n\n  for (const element of data) {\n    const key = element['s2_id']\n    const metric = element[visibleMetric]\n    if (newCells.has(key)) {\n      const polygon = newCells.get(key)\n      polygon['uniq_users'] = element['uniq_users']\n      polygon['signal_avg'] = element['signal_avg']\n      polygon.setOptions({\n        strokeColor: colorScale(metric),\n        fillColor: colorScale(metric),\n      })\n      continue\n    }\n    const paths = []\n    for (const coordinate of element['s2_coordinates']) {\n      paths.push({ lat: coordinate.lat, lng: coordinate.lon })\n    }\n    const polygon = new google.maps.Polygon({\n      paths: paths,\n      strokeColor: colorScale(metric),\n      strokeOpacity: 0.8,\n      strokeWeight: 2,\n      fillColor: colorScale(metric),\n      fillOpacity: 0.35,\n    })\n    polygon['s2_id'] = element['s2_id']\n    polygon['uniq_users'] = element['uniq_users']\n    polygon['signal_avg'] = element['signal_avg']\n    newCells.set(key, polygon)\n    polygon.setMap(map)\n    polygon.addListener('click', handleCellClick)\n  }\n  cells = newCells\n}\n\nasync function fetchData (map) {\n  const mapBounds = map.getBounds()\n  const request = {\n    area: [\n      {\n        lat: mapBounds.getSouthWest().lat(),\n        lon: mapBounds.getSouthWest().lng()\n      },\n      {\n        lat: mapBounds.getNorthEast().lat(),\n        lon: mapBounds.getNorthEast().lng()\n      }\n    ]\n  }\n  const response = await fetch('http://127.0.0.1:8080/data', {\n    method: 'POST',\n    headers: {\n      'Accept': 'application/json',\n      'Content-Type': 'application/json'\n    },\n    body: JSON.stringify(request)\n  })\n  return await response.json()\n}\n\nfunction clearNotVisibleCells (data) {\n  const newCells = new Map()\n  for (const element of data) {\n    const key = element['s2_id']\n\n    if (cells.has(key)) {\n      newCells.set(key, cells.get(key))\n    }\n  }\n  for (const [key, value] of cells) {\n    if (!newCells.has(key)) {\n      value.setMap(null)\n    }\n  }\n  return newCells\n}\n\nfunction getColorScale (data) {\n  return d3.scaleSequential()\n    .interpolator(d3.interpolateWarm)\n    .domain([d3.min(data.map((e) => e[visibleMetric])), d3.max(data.map((e) => e[visibleMetric]))])\n}\n\nfunction handleCellClick (e) {\n  if (infoWindow === null) {\n    infoWindow = new google.maps.InfoWindow({})\n  }\n  infoWindow.close()\n  let metricValue = this[visibleMetric]\n  if (metricValue % 1 !== 0) {\n    metricValue = metricValue.toFixed(2)\n  }\n  infoWindow.setContent(`\n    <b>${this['s2_id'].toString(16)}</b>\n    <br/>\n    <div class=\"center-text\">${metricValue}</div>\n`)\n  infoWindow.setPosition(e.latLng)\n  infoWindow.open(this.map)\n}\n\n"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "style.css",
		FileModTime: time.Unix(1623911444, 0),

		Content: string("#map {\n    height: 100%;\n}\n\nhtml,\nbody {\n    height: 100%;\n    margin: 0;\n    padding: 0;\n}\n\n.center-text {\n    text-align: center;\n}\n\n.radio-control {\n    background-color: #fff;\n    border: 2px solid #fff;\n    border-radius: 3px;\n    box-shadow: 0 2px 6px rgba(0, 0, 0, .3);\n    margin-top: 8px;\n    margin-bottom: 22px;\n    text-align: center;\n}\n\n.radio-item {\n    color: #191919;\n    font-family: Roboto, Arial, sans-serif;\n    font-size: 16px;\n    line-height: 38px;\n    padding-left: 5px;\n    padding-right: 5px;\n}"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1624876502, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "index.html"
			file3, // "index.js"
			file4, // "style.css"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`../../static`, &embedded.EmbeddedBox{
		Name: `../../static`,
		Time: time.Unix(1624876502, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"index.html": file2,
			"index.js":   file3,
			"style.css":  file4,
		},
	})
}