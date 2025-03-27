package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"go.avito.ru/av/service-ratings-users-composition/pkg/linters/errchecklog"
)

func main() {
	singlechecker.Main(errchecklog.Analyzer)
}
