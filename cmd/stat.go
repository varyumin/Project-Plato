/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"
	"text/tabwriter"
)

var (
	URLs []string
)

type Answer struct {
	URL string
	BodySize int64
	StatusCode int
}
type Answers []Answer

func (a Answers) Len() int {
	return len(a)
}

func (a Answers) Less(i, j int) bool {
	return a[i].BodySize > a[j].BodySize
}

func (a Answers) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func Request (u string) (a Answer){
	h, err := url.Parse(u)
	if err != nil {
		log.Errorln(err)
		return Answer{}
	}

	client := &http.Client {}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Errorln(err)
		return Answer{}
	}
	res, err := client.Do(req)
	if err != nil {
		log.Errorln(err)
		return Answer{}
	}
	defer res.Body.Close()

	b, err := io.Copy(ioutil.Discard, res.Body)
	if err != nil {
		log.Errorln(err)
		return Answer{}
	}
	a.BodySize = b
	a.URL = h.Host
	a.StatusCode = res.StatusCode
return a
}



// statCmd represents the stat command
var statCmd = &cobra.Command{
	Use:   "stat",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var as Answers

		wg := &sync.WaitGroup{}
		mutex := &sync.Mutex{}

		for _, url := range URLs {
			wg.Add(1)

			go func(s string) {
				if u := Request(s); u != (Answer{}) {
					mutex.Lock()
					as = append(as, u)
					mutex.Unlock()
				}
				wg.Done()
			}(url)
		}
	wg.Wait()
	sort.Sort(as)
	PrintResult(as)
	},
}

func PrintResult(a Answers)  {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, ' ', 0)
	defer w.Flush()
	fmt.Fprintf(w, "\n %s\t\t%s\t\t%s\t", "Host", "Body size(byte)", "Status Code")
	fmt.Fprintf(w, "\n %s\t\t%s\t\t%s\t", "----", "----", "----")

	for _, u := range a{
		fmt.Fprintf(w, "\n %s\t\t%d\t\t%d\t", u.URL, u.BodySize, u.StatusCode)
	}
	fmt.Fprint(w, "\n")
}

func init() {
	rootCmd.AddCommand(statCmd)
	statCmd.Flags().StringSliceVarP(&URLs, "urls", "u", nil, "Take a list of URLs ")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
