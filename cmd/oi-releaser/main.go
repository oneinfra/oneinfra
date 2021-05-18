/**
 * Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 **/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/oneinfra/oneinfra/internal/app/oi-releaser/text"
)

func main() {
	app := &cli.App{
		Usage: "oneinfra releaser CLI tool",
		Commands: []*cli.Command{
			{
				Name:  "text",
				Usage: "text operations",
				Subcommands: []*cli.Command{
					{
						Name:  "replace-placeholders",
						Usage: "replace placeholders on text provided through stdin and print result to stdout",
						Action: func(c *cli.Context) error {
							stdin, err := ioutil.ReadAll(os.Stdin)
							if err != nil {
								return err
							}
							fmt.Print(text.ReplacePlaceholders(string(stdin)))
							return nil
						},
					},
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
