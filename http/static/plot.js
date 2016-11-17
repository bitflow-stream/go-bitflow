
PLOT_WIDTH = 800
PLOT_HEIGHT = 300
PAUSED = false

function update_all_metrics_individually() {
	d3.json("/metrics", function(metrics){
		$.each(metrics, function(i, name) {
			update_one_metric(name)
		})
	})
}

function update_one_metric(name) {
	d3.json("/data", function(data) {
		update_metric_data(name, data)
	})
}

function update_all_metrics() {
	d3.json("/data", function(allMetrics) {
		$.each(allMetrics, function(name, data) {
			update_metric_data(name, data)
		})
	})
}

function update_metric_data(name, data) {
	plotData = []
	$.each(data, function(i, val){
		plotData[i] = { index: i, value: val }
	})
	// 'date':new Date('2014-11-01')

	target = get_target_div(name)
	MG.data_graphic({
		title: name,
		data: plotData,
		width: PLOT_WIDTH,
		height: PLOT_HEIGHT,
		target: target,
		color: randomColor({luminosity: 'dark'}),
		x_accessor: 'index',
		y_accessor: 'value',
		linked: true,
	})
}

function stringHash(s) {
	var hash = 0, i, chr, len;
	if (s.length === 0) return hash;
	for (i = 0, len = s.length; i < len; i++) {
		chr   = s.charCodeAt(i);
		hash  = ((hash << 5) - hash) + chr;
		hash |= 0; // Convert to 32bit integer
	}
	return hash;
}

function get_target_div(name) {
	hash = stringHash(name)
	id = "__data__" + hash
	sel = "#" + id
	if ($(sel).length == 0) {
		code = '<div id="' + id + '" class="data_plot"></div>'
		$('.data_container').append(code);

		$(sel)
		.resizable({
			alsoResize: ".data_plot",
			grid: 100,
			resize: function(event, ui) {
				PLOT_HEIGHT = ui.size.height
				PLOT_WIDTH = ui.size.width
			},
		})
		.click(toggle_pause);
	}
	return sel
}

function toggle_pause() {
	PAUSED = !PAUSED
}

function loop_update_metrics() {
	if (!PAUSED) update_all_metrics()
	setTimeout(loop_update_metrics, 1000)
}

$(loop_update_metrics)
