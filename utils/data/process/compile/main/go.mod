module generate-results-db

go 1.18

require FairFareFinder/src/go_files/timeutils v0.0.0

require (
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/schollz/progressbar/v3 v3.14.3 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/term v0.20.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace FairFareFinder/src/go_files/timeutils => ../../../src/go_files/timeutils