module github.com/Tris20/FairFareFinder

go 1.23.1

require (
	github.com/Tris20/FairFareFinder/src/backend v0.0.1
	github.com/Tris20/FairFareFinder/utils/time-and-date v0.0.1
	github.com/gorilla/sessions v1.4.0
	github.com/mattn/go-sqlite3 v1.14.23
	github.com/schollz/progressbar/v3 v3.14.2
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/term v0.17.0 // indirect
)

replace github.com/Tris20/FairFareFinder/utils/time-and-date => ./utils/time-and-date

replace github.com/Tris20/FairFareFinder/src/backend => ./src/backend
