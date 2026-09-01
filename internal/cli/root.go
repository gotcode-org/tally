/*
Copyright 2026 The GotCode Collective

This file is part of Tally.

Tally is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/

package cli

import (
	"fmt"
	"os"
)

// Execute is the entrypoint for the Cobra CLI root command.
func Execute() {
	fmt.Println("Tally: The Time-Tracking Command Line")
	os.Exit(0)
}
