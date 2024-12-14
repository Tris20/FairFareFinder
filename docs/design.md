# Overview

# Components

## Main

the main component (FairFareFinder main) is the entry point of the application. It is responsible for parsing the command line arguments, reading the configuration file, and starting the server.

##

# Operation

# Error Handling

After running the application `api.log` file will be created in the root directory. Check out the log for details about what may have gone wrong.

## SSH to server

You need to be able to ssh into the server to get the database and images. For this you need to make an ssh key and add it to the server.

<!-- we will update this when the next person needs to do it -->

```bash
ssh-keygen -t rsa -b 4096 -C "
```

The key needs to be added to the server. Go go

## Database missing

```text
2024/12/14 12:46:54 Error querying cities: no such table: flight
```

If you are getting any error related to this, make sure to get a copy of the database from the server.
The file for the sqlite database should be here:

```
"./data/compiled/main.db"
```

fix this by copying the database file to the correct location.

```bash
cd FairFareFinder
scp root@fairfarefinder.com:~/FairFareFinder/data/compiled/main.db ./data/compiled/main.db
```

and back it up

```bash
cp ./data/compiled/main.db ./data/compiled/main.db.bak
```

## images missing

```bash
mkdir -p ./ignore/location-images
rsync -avz -e "ssh" root@fairfarefinder.com:~/FairFareFinder/ignore/location-images/ ./ignore/location-images/
```
