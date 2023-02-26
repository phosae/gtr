package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

func main() {
	var config string
	var values string
	var debug bool
	flag.StringVar(&config, "c", "", "config(template) file name")
	flag.StringVar(&values, "v", "", "values file name")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.Parse()

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var cfgBytes []byte
	if config == "-" {
		if cfgBytes, err = io.ReadAll(os.Stdin); err != nil {
			log.Fatal(fmt.Errorf("unable to read config from std: %v", err))
		}
	} else {
		cfgBytes, err = readFileOrDefaultFromPwd(pwd, config, []string{"cfg", "config", "c"})
		if err != nil || len(cfgBytes) == 0 {
			log.Fatal(fmt.Errorf("empty config or fatal readfile/%v", err))
			return
		}
	}

	valBytes, err := readFileOrDefaultFromPwd(pwd, values, []string{"vars", "vars.yml", "vars.yaml", "vars.json",
		"values", "values.yml", "values.yaml", "values.json"})
	if err != nil {
		log.Fatal(fmt.Errorf("unable to read config variables: %v", err))
	}

	if debug {
		fmt.Printf("config: \n%s\n", string(cfgBytes))
		fmt.Printf("values: \n%s\n", string(valBytes))
	}

	ret, err := render(cfgBytes, func() (map[string]interface{}, error) {
		var varMap = map[string]interface{}{}
		if envVars := os.Getenv("V"); envVars != "" {
			for _, pair := range strings.Split(envVars, ",") {
				kv := strings.Split(strings.TrimSpace(pair), "=")
				if len(kv) != 2 {
					log.Fatal(fmt.Errorf("invalid value env: %v, valid format `k1=v1,k2=v2`", err))
				}
				varMap[kv[0]] = kv[1]
			}
			return varMap, nil
		}

		d2 := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(valBytes), 1024)
		if err = d2.Decode(&varMap); err != nil {
			return nil, fmt.Errorf("only json/yaml values was supported")
		}
		return varMap, nil
	})
	if err != nil {
		log.Fatal(fmt.Errorf("err render cfg: %v", err))
	}
	fmt.Println(string(ret))
}

func readFileOrDefaultFromPwd(pwd, file string, defaults []string) ([]byte, error) {
	if filepath.IsAbs(file) {
		return os.ReadFile(file)
	} else if file != "" {
		return os.ReadFile(filepath.Join(pwd, file))
	} else {
		for _, name := range defaults {
			detectVal := filepath.Join(pwd, name)
			if _, err := os.Stat(detectVal); err == nil {
				return os.ReadFile(detectVal)
			}
		}
	}
	return nil, nil
}

func render(template []byte, tovars func() (map[string]interface{}, error)) (rendered []byte, err error) {
	varMap, err := tovars()
	if err != nil {
		return nil, fmt.Errorf("err got vars map: %v", err)
	}

	if len(template) == 0 || len(varMap) == 0 {
		return template, nil
	}
	return merge(template, varMap)
}

func merge(cfgTmplate []byte, vars map[string]interface{}) ([]byte, error) {
	var gtm = template.New("config_render").Funcs(funcMap()).Option("missingkey=default")
	tp, err := gtm.Parse(string(cfgTmplate))
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err := tp.Execute(&out, vars); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
