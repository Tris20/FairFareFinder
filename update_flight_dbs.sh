#!/bin/bash

#### Update the Flight Schedule ###
# Change to the aerodata directory
cd ~/Documents/Workspace/SVN/SVN_BASE/Software/Shared_Projects/FairFareFinder/utils/database_utils/flights/update-schedule/aerodatabox/

# Run the aerodata script
./aerodatabox


### Update the flight prices of the new schedule ####
# Change directory
cd ~/Documents/Workspace/SVN/SVN_BASE/Software/Shared_Projects/FairFareFinder || exit

# Update flight prices
./FairFareFinder updateFlightPrices
if [ $? -ne 0 ]; then
    echo "Failed to update flight prices"
    exit 1
fi


### Send the DB with the new flight info over to the web server ###
# Copy the flights.db to local backup directory
scp -i ~/.ssh/fff_server root@fairfarefinder.com:~/FairFareFinder/data/flights.db /home/tristan/Documents/Workspace/SVN/SVN_BASE/Software/Shared_Projects/Backups/FairFareFinder/
if [ $? -ne 0 ]; then
    echo "Failed to copy flights.db to backup"
    exit 1
fi

# Copy the updated flights.db back to the server's incoming directory
scp -i ~/.ssh/fff_server /home/tristan/Documents/Workspace/SVN/SVN_BASE/Software/Shared_Projects/FairFareFinder/data/flights.db root@fairfarefinder.com:~/FairFareFinder/data/incoming_db/flights.db
if [ $? -ne 0 ]; then
    echo "Failed to copy flights.db to server"
    exit 1
fi

echo "Operations completed successfully"

