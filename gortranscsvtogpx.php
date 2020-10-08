<?php

$inDir = 'data_gortrans/';
$outDir = 'gpx/';

foreach($argv as $arg) {
	list($key, $val) = explode('=', $arg);
	
	if ($key == '-indir') $inDir = $val;
	if ($key == '-outdir') $outDir = $val;
	if ($key == '-from') $startTime = $val;
	if ($key == '-to') $endTime = $val;
}

echo "Сканирование каталога: " . $inDir . PHP_EOL;

$files = array_slice(scandir($inDir), 2);
$files = array_filter($files, function($k, $v) use ($startTime,$endTime) {
		$res = strpos($k, 'response-data-') !== false;
		if (isset($startTime)) $res = $res && $k >= 'response-data-' . date('Y-m-d', strtotime($startTime)) . '.csv';
		if (isset($endTime)) $res = $res && $k <= 'response-data-' . date('Y-m-d', strtotime($endTime)) . '.csv';
		return $res;
	}, ARRAY_FILTER_USE_BOTH);
echo "Найдено файлов: " . count($files) . PHP_EOL;

if (isset($startTime)) $startDt = DateTime::createFromFormat('Y-m-d H:i:s', $startTime);
if (isset($endTime)) $endDt = DateTime::createFromFormat('Y-m-d H:i:s', $endTime);
		
$vehicles = [];
foreach($files as $filename) {
	$csv = file($inDir . $filename);
	foreach ($csv as $line) {
		$csvLine = str_getcsv(trim($line), ';');
		$dt = DateTime::createFromFormat('Y-m-d H:i:s', str_replace('T', ' ', $csvLine[0]));
		if ((isset($startDt) && $dt < $startDt) || (isset($endDt) && $dt > $endDt)) continue;
		$vehicles[$csvLine[1]]['name'] = $csvLine[3] . ' ' . $csvLine[1];
		$vehicles[$csvLine[1]]['track'][] = [
			'time' => $dt->format("Y-m-d\TH:i:s\Z"),
			'lng' => $csvLine[5],
			'lat' => $csvLine[6]
		];
		
	}
}

echo "Найдено транспортных средств: " . count($vehicles) . PHP_EOL;

if (!file_exists($outDir)) {
	mkdir($outDir);
}

foreach($vehicles as $vehicle) {
	$gpx = '<?xml version="1.0" encoding="UTF-8"?>
	<gpx creator="gortrans" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd" version="1.1" xmlns="http://www.topografix.com/GPX/1/1">
	 <metadata>
	  <time>' . $vehicle["track"][0]["time"] . '</time>
	 </metadata>';
	
	$gpx .= startTrk($vehicle["name"]);
	
	$prevPoint = null;
	foreach($vehicle["track"] as $point) {
		if ($prevPoint != null && distance($prevPoint, $point) < 0.0000001) {
			$prevPoint = $point;
			continue;
		}
		if ($prevPoint != null && distance($prevPoint, $point) > 0.01) {
			$gpx .= endTrk();
			$gpx .= startTrk($vehicle["name"]);
		}
		$gpx .= addPoint($point);
		$prevPoint = $point;
	}
	$gpx .= endTrk();
	$gpx .= '</gpx>';
	
	file_put_contents($outDir . $vehicle["name"] . ".gpx", $gpx);
}

echo "GPX файлы созданы" . PHP_EOL;





function startTrk($name) {
	return '<trk><name>' . $name . '</name><type>1</type><trkseg>';
}

function endTrk() {
	return '</trkseg></trk>';
}

function addPoint($point) {
	return '<trkpt lat="' . $point["lat"] . '" lon="' . $point["lng"] . '"><time>' . $point["time"] . '</time></trkpt>';
}

function distance($point1, $point2) {
	return sqrt(pow($point1['lat']-$point2['lat'], 2)+pow($point1['lng']-$point2['lng'], 2));
}

function speed($point1, $point2) {
	return sqrt(pow($point1['lat']-$point2['lat'], 2)+pow($point1['lng']-$point2['lng'], 2))/(strtotime($point2['time'])-strtotime($point1['time']));
}