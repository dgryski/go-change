package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/dgryski/go-change/offline"
)

func main() {
	minCorrelation := flag.Float64("c", 0.6, "correlation threshold")
	markerWidth := flag.Int("w", 1, "marker width")
	fname := flag.String("f", "", "file name")
	ymin := flag.Int("ymin", 0, "minimum y value for graph")

	flag.Parse()

	var f io.Reader

	if *fname == "" {
		log.Println("reading from stdin")
		f = os.Stdin
	} else {
		var err error
		f, err = os.Open(*fname)
		if err != nil {
			fmt.Println("open failed:", err)
			return
		}
	}

	scanner := bufio.NewScanner(f)

	var series []float64
	var graphData [][]float64
	var cnt int

	for scanner.Scan() {
		value, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			fmt.Printf("error parsing <%s>: %s\n", scanner.Text(), err)
			continue
		}

		series = append(series, value)
		graphData = append(graphData, []float64{float64(cnt), value})

		cnt += 1
	}

	detector := offline.Detector{*markerWidth, *minCorrelation}
	changes, err := detector.Check(series)
	if err != nil {
		log.Fatal(err)
	}

	var changePoints []int
	for _, change := range changes {
		changePoints = append(changePoints, change.Index)

		log.Printf("Found change at pos=%d with corr=%.4f", change.Index, change.Correlation)
	}

	reportTmpl.Execute(os.Stdout, struct {
		YMin         int
		GraphData    [][]float64
		ChangePoints []int
	}{
		*ymin,
		graphData,
		changePoints,
	})
}

var reportTmpl = template.Must(template.New("report").Parse(`
<html>
<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.2/jquery.flot.min.js"></script>

<script type="text/javascript">

    var data = {{ .GraphData }};

    $(document).ready(function() {
        $.plot($("#placeholder"), [data], {
             yaxis: { min: {{ .YMin }} },
             grid: {
                markings: [
                  {{ range .ChangePoints }}{ color: '#000', lineWidth: 1, xaxis: { from: {{ . }}, to: {{ . }} } },
                  {{ end }}
                ]
              }
           })
        })

</script>

<body>

<div id="placeholder" style="width:1200px; height:400px"></div>

</body>
</html>
`))
