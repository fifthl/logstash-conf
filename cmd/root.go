package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&conf, "conf", "", "", "*specify logstash.conf file path")
	rootCmd.PersistentFlags().StringVarP(&topic, "topic", "", "", "*kafka topic name, example: test.abc")
	rootCmd.PersistentFlags().StringVarP(&group, "group", "", "donet", "specify kafka group id")

}

var (
	//go:embed templates/input.template
	inputContent string
	//go:embed templates/output.template
	outputContent string
)

var (
	conf     string
	topic    string
	group    string
	inputRe  = regexp.MustCompile(`(?m)(?s)(?P<input>input\s*\{)|(?P<kafka>kafka\s*\{(.*?)\})`)
	outputRe = regexp.MustCompile(`(?m)(?s)(?P<output>output\s*\{)|(?P<if>if\s*\[(.*?)\}\s*})`)

	// 全局 Cmd
	rootCmd = &cobra.Command{
		Use:     "logstash-conf",
		Example: "logstash-conf [--conf] [--topic] [--group]",
		Run:     RunFunc,
	}
)

type Kafka struct {
	Topic   string
	Index   string
	GroupId string
}

func formatData(topic, group, conf string) Kafka {

	if topic == "" {
		fmt.Println("topic not specified, --topic parameter missing")
		os.Exit(1)
	}

	if conf == "" {
		fmt.Println("configuration file path not specified, --conf parameter missing")
		os.Exit(1)
	}

	if cont := strings.Contains(topic, "."); !cont == true {
		fmt.Println(topic)
		fmt.Println("topic format is incorrect, sample value: test.abc")
		os.Exit(1)
	}

	i := strings.Split(topic, ".")
	index := i[0] + i[1]

	data := Kafka{
		Topic:   topic,
		Index:   index,
		GroupId: group,
	}

	return data
}

// 生成 kafka 阶段配置
func generateInputStage(data Kafka) string {
	var tpl bytes.Buffer

	tmpl, _err := template.New("inputTmpl").Parse(inputContent)
	if _err != nil {
		log.Fatalf("template parse error: %v", _err)
	}

	// 将传入的flag写入模板
	if err := tmpl.Execute(&tpl, data); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return tpl.String() + "\n}"
}

// 生成 input 阶段配置
func generateOutputStage(data Kafka) string {
	tmpl, _err := template.New("outputTmpl").Parse(outputContent)
	if _err != nil {
		log.Fatalf("template parse error: %v", _err)
	}

	var tpl bytes.Buffer

	// 将传入的flag写入模板
	if err := tmpl.Execute(&tpl, data); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return tpl.String() + "\n}"
}

// 主函数
func RunFunc(cmd *cobra.Command, args []string) {
	/*
		1. 利用正则匹配 kafka 和 output 的每个阶段
		2. 利用传递的 flag 生成相应阶段的模板
		3. 生成的模板追加到kafka或者output切片的尾部
		4. 重写文件
	*/

	// 将 flag 格式化
	flagsData := formatData(topic, group, conf)

	input := generateInputStage(flagsData)
	output := generateOutputStage(flagsData)

	dataByte, err := os.ReadFile(conf)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	// input 匹配
	inputStage := inputRe.FindAllString(string(dataByte), -1)
	inputStage = append(inputStage, input)

	//output 匹配
	outputStage := outputRe.FindAllString(string(dataByte), -1)
	outputStage = append(outputStage, output)

	//格式配置文件
	outputStage[0] = "\n" + outputStage[0]
	outputStage[1] = "\n  " + outputStage[1]
	inputStage[1] = "\n   " + inputStage[1]

	// 两个阶段合并
	inputStage = append(inputStage, outputStage...)

	// 写入
	f, _err := os.OpenFile(conf, os.O_RDWR, 0755)
	if _err != nil {
		fmt.Printf("%v\n", _err)
		os.Exit(1)
	}
	defer f.Close()

	f.Truncate(0)
	f.Seek(0, 0)

	var b []byte
	for _, str := range inputStage {
		b = append(b, str...)
	}

	if err = os.WriteFile(conf, b, 0755); err != nil {
		fmt.Println("write file err ", err)
		os.Exit(1)
	}
	fmt.Printf("%v add done\n", flagsData.Topic)
	os.Exit(0)
}

func Execute() {
	rootCmd.Execute()
}
