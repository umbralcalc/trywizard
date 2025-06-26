package features

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

func NewDataFrameScatter(
	df *dataframe.DataFrame,
	xAxis string,
	yAxis string,
	groupBy ...string,
) *charts.Scatter {
	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		charts.WithXAxisOpts(opts.XAxis{Name: xAxis}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger:   "item",
			Formatter: "({c})",
		}),
	)

	groupCol := ""
	if len(groupBy) > 0 {
		groupCol = groupBy[0]
	}

	if groupCol != "" {
		// Automatically get unique group values
		groups := df.Col(groupCol).Records()
		uniqueGroups := make(map[string]bool)
		for _, g := range groups {
			uniqueGroups[g] = true
		}

		for group := range uniqueGroups {
			points := make([]opts.ScatterData, 0)
			fdf := df.Filter(dataframe.F{
				Colname:    groupCol,
				Comparator: series.Eq,
				Comparando: group,
			})

			xSeries := fdf.Col(xAxis)
			ySeries := fdf.Col(yAxis)
			for i := 0; i < xSeries.Len(); i++ {
				points = append(points, opts.ScatterData{
					Value: []interface{}{
						xSeries.Elem(i).Val(),
						ySeries.Elem(i).Val(),
					},
				})
			}

			scatter.AddSeries(group, points)
		}
	} else {
		// No grouping, single series
		points := make([]opts.ScatterData, 0)
		xSeries := df.Col(xAxis)
		ySeries := df.Col(yAxis)
		for i := 0; i < xSeries.Len(); i++ {
			points = append(points, opts.ScatterData{
				Value: []interface{}{
					xSeries.Elem(i).Val(),
					ySeries.Elem(i).Val(),
				},
			})
		}
		scatter.AddSeries(yAxis, points)
	}

	return scatter
}

func NewDataFrameLine(
	df *dataframe.DataFrame,
	xAxis string,
	yAxis string,
	groupBy ...string,
) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithXAxisOpts(opts.XAxis{Name: xAxis}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger:   "item",
			Formatter: "({c})",
		}),
	)

	groupCol := ""
	if len(groupBy) > 0 {
		groupCol = groupBy[0]
	}

	if groupCol != "" {
		// Automatically get unique group values
		groups := df.Col(groupCol).Records()
		uniqueGroups := make(map[string]bool)
		for _, g := range groups {
			uniqueGroups[g] = true
		}

		for group := range uniqueGroups {
			points := make([]opts.LineData, 0)
			fdf := df.Filter(dataframe.F{
				Colname:    groupCol,
				Comparator: series.Eq,
				Comparando: group,
			})

			xSeries := fdf.Col(xAxis)
			ySeries := fdf.Col(yAxis)
			for i := 0; i < xSeries.Len(); i++ {
				points = append(points, opts.LineData{
					Value: []interface{}{
						xSeries.Elem(i).Val(),
						ySeries.Elem(i).Val(),
					},
				})
			}

			line.AddSeries(group, points)
		}
	} else {
		// No grouping, single series
		points := make([]opts.LineData, 0)
		xSeries := df.Col(xAxis)
		ySeries := df.Col(yAxis)
		for i := 0; i < xSeries.Len(); i++ {
			points = append(points, opts.LineData{
				Value: []interface{}{
					xSeries.Elem(i).Val(),
					ySeries.Elem(i).Val(),
				},
			})
		}
		line.AddSeries(yAxis, points)
	}

	return line
}
