/*
 * Copyright (c) 2019 PANTHEON.tech.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"log"
	"os"

	"github.com/PantheonTechnologies/vpptop/gui"
	"github.com/PantheonTechnologies/vpptop/stats"
)

var (
	statSock = flag.String("socket", stats.DefaultSocket, "VPP stats segment socket")
	logFile  = flag.String("log", "vpptop.log", "Log file")
)

var (
	lightTheme = false
)

func main() {
	flag.Parse()

	if _, lightTheme = os.LookupEnv("VPPTOP_THEME_LIGHT"); lightTheme {
		gui.SetLightTheme()
	}

	logs, err := os.Create(*logFile)
	if err != nil {
		log.Fatalf("error occured while creating file: %v", err)
	}
	defer logs.Close()

	app := NewApp()
	if err := app.Init(*statSock); err != nil {
		log.Fatalf("error occured during init: %v", err)
	}
	log.SetOutput(logs)

	app.Run()
}
