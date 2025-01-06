package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v2"
)

type APIServerMetrics struct {
	APIServerMetrics map[string][]MetricsValue
}

type YAMLMetricsData struct {
	Name   string
	Path   string
	Type   string
	Help   string
	Labels map[string]string
	Values map[string]string
}

type MetricsValue struct {
	Metric interface{}   `json:"metric"`
	Value  []interface{} `json:"value"`
}

var parseCommand = &cli.Command{
	Name:  "parse",
	Usage: "parse a clusterloader2 json output to yaml for prometheus exporter",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "input",
			Usage: "input file",
		},
		&cli.StringFlag{
			Name:  "output",
			Usage: "output file",
		},
	},
	Action: parse,
}

func parse(ctx context.Context, cmd *cli.Command) error {
	input := ""
	if cmd.NArg() > 0 {
		input = cmd.Args().Get(0)
	}
	jsonFile, err := os.Open(input)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var rawMetrics APIServerMetrics

	err = json.Unmarshal([]byte(byteValue), &rawMetrics)
	if err != nil {
		fmt.Println(err)
		return err
	}

	metrics := rawMetrics.APIServerMetrics

	yamlMetrics := []YAMLMetricsData{}

	for metricName := range metrics {
		data := YAMLMetricsData{
			Name: metricName,
			Path: "{ .APIServerMetrics." + metricName + "[*] }",
			Help: "help",
			Type: "object",
			Labels: map[string]string{
				"metadata": "{.metric}",
			},
			Values: map[string]string{
				"value": "{.value[1]}",
			},
		}
		yamlMetrics = append(yamlMetrics, data)
	}

	result := map[string]interface{}{
		"modules": map[string]interface{}{
			"default": map[string]interface{}{
				"metrics": yamlMetrics,
			},
		},
	}

	yamlData, err := yaml.Marshal(result)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(string(yamlData))

	return nil
}
