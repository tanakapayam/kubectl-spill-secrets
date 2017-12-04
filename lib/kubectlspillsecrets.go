package kubectlspillsecrets

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/fatih/color"
	flag "github.com/ogier/pflag"
	"gopkg.in/yaml.v2"
)

const (
	README  = "README.md"
	VERSION = "VERSION.txt"
)

var (
	bold                   = color.New(color.Bold).SprintFunc()
	ejsonPublicKey         = ""
	hyphenToUnderscoreKeys = false
	redacted               = false
	uppercaseKeys          = false
)

type Secret struct {
	APIVersion string
	Data       map[string]string
	Kind       string
	Metadata   struct {
		Name      string
		Namespace string
	}
	Type string
}

func ParseArgs() {
	var h2 = regexp.MustCompile(`^#+ (.+)$`)
	var indented = regexp.MustCompile(`^[ \t]`)

	flag.Usage = func() {
		_, program, _, _ := runtime.Caller(0)

		// read in README.md
		if readme, err := os.Open(
			path.Dir(program) + "/../" + README,
		); err == nil {
			// make sure it gets closed
			defer readme.Close()

			// create a new scanner and read the file line by line
			scanner := bufio.NewScanner(readme)
			for scanner.Scan() {
				// check for #
				h2Result := h2.FindStringSubmatch(scanner.Text())
				if h2Result != nil {
					fmt.Printf("%s\n", bold(h2Result[1]))
				} else {
					indentedResult := indented.FindStringSubmatch(scanner.Text())
					if indentedResult == nil {
						fmt.Println("    " + scanner.Text())
					} else {
						fmt.Println(scanner.Text())
					}
				}
			}

			// check for errors
			if err = scanner.Err(); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}

		// read in VERSION.txt
		if version, err := ioutil.ReadFile(
			path.Dir(program) + "/../" + VERSION,
		); err == nil {
			fmt.Printf("\n%s\n    %s\n", bold("VERSION"), string(version))
		} else {
			log.Fatal(err)
		}

		fmt.Printf("%s\n", bold("FLAGS"))
		flag.PrintDefaults()
	}
	flag.StringVar(&ejsonPublicKey, "ejson-public-key", "", "outputs ejson-formatted object with 64-character public key")
	flag.BoolVar(&hyphenToUnderscoreKeys, "hyphen-to-underscore-keys", false, "replaces hyphens in data keys to underscores")
	flag.BoolVar(&redacted, "redacted", false, `sets data values to "--REDACTED--"`)
	flag.BoolVar(&uppercaseKeys, "uppercase-keys", false, "uppercases data keys")
	flag.Parse()

	if len(ejsonPublicKey) != 0 && len(ejsonPublicKey) != 64 {
		flag.Usage()
		os.Exit(2)
	}
}

func transformData(data map[string]string) {
	// sort data keys
	var keys []string

	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	i := 0
	comma := ","
	for _, k := range keys {
		i++
		if i == len(keys) {
			comma = ""
		}

		v := ""
		if redacted {
			v = `--REDACTED--`
		} else {
			v2, err := base64.StdEncoding.DecodeString(data[k])
			if err != nil {
				log.Fatal(err)
			}
			v = strings.TrimRight(string(v2), "\n")
		}
		if hyphenToUnderscoreKeys {
			k = strings.Replace(k, "-", "_", -1)
		}
		if uppercaseKeys {
			k = strings.ToUpper(k)
		}

		if ejsonPublicKey == "" {
			fmt.Println("  " + k + ": " + v)
		} else {
			fmt.Println(`        "` + k + `": "` + v + `"` + comma)
		}
	}
}

func SpillSecrets() {
	r := regexp.MustCompile(`[a-z][a-zA-Z]+: `)

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	sData := r.ReplaceAllStringFunc(string(data), func(m string) string {
		return strings.ToLower(m)
	})

	secret := Secret{}

	err = yaml.Unmarshal([]byte(sData), &secret)
	if err != nil {
		log.Fatal(err)
	}

	// update secret properties
	if secret.Metadata.Name == "app-secrets" {
		secret.Metadata.Name = "secrets"
	}
	if secret.Type == "string" {
		secret.Type = "Opaque"
	}

	if ejsonPublicKey == "" {
		fmt.Println("apiVersion: " + secret.APIVersion)
		fmt.Println("data:")
		transformData(secret.Data)
		fmt.Println("kind: " + secret.Kind)
		fmt.Println("metadata:")
		fmt.Println("  name: " + secret.Metadata.Name)
		fmt.Println("  namespace: " + secret.Metadata.Namespace)
		fmt.Println("type: " + secret.Type)
	} else {
		fmt.Println(`{`)
		fmt.Println(`  "_public_key": "` + ejsonPublicKey + `",`)
		fmt.Println(`  "kubernetes_secrets": {`)
		fmt.Println(`    "app-secrets": {`)
		fmt.Println(`      "_type": "` + secret.Type + `",`)
		fmt.Println(`      "data": {`)
		transformData(secret.Data)
		fmt.Println(`      }`)
		fmt.Println(`    }`)
		fmt.Println(`  }`)
		fmt.Println(`}`)
	}
}
