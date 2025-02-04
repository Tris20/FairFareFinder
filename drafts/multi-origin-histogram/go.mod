module draft-multi-origin

go 1.23.1

require (
	github.com/Tris20/FairFareFinder/src/backend v0.0.1
	github.com/Tris20/FairFareFinder/src/backend/config v0.0.1
	github.com/Tris20/FairFareFinder/src/backend/model v0.0.1
	github.com/mattn/go-sqlite3 v1.14.23
)

require (
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/gorilla/sessions v1.4.0 // indirect
)

replace github.com/Tris20/FairFareFinder/src/backend => ../../src/backend

replace github.com/Tris20/FairFareFinder/src/backend/model => ../../src/backend/model

replace github.com/Tris20/FairFareFinder/src/backend/config => ../../src/backend/config
