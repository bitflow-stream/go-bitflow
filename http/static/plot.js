
function update_all_metrics() {
	$.ajax("/metrics", {
		async: false,
		dataType: "json",
		success: function(metrics, status) {
			$.each(metrics, function(i, name) {
				update_metric(name)
			})
		}
	})
}

function update_metric(name) {
	$.ajax("/data", {
		data: "metric=" + name,
		async: false,
		dataType: "json",
		success: function(data, status) {
			plotData = []
			$.each(data, function(i, val){
				plotData[i] = { index: i, value: val }
			})
			// 'date':new Date('2014-11-01')

			target = get_target_div(name)
			MG.data_graphic({
				title: name,
				// description: "...",
				data: plotData,
				width: 800,
				height: 300,
				target: target,
				x_accessor: 'index',
				y_accessor: 'value',
			})
		}
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
};

function get_target_div(name) {
	hash = stringHash(name)
	id = "__data__" + hash
	sel = "#" + id
	if ($(sel).length == 0) {
		code = '<div id="' + id + '" class="data_plot"></div>'
		$('.data_container').append(code);
	}
	return sel
}

function loop_update_metrics() {
	update_all_metrics()
	setTimeout(loop_update_metrics, 1000)
}

$(loop_update_metrics)
