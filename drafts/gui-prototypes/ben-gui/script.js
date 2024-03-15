document.getElementById('spawnBoxButton').addEventListener('click', spawnBox);

//this spawns a box
function spawnBox(){
    const newBox = new LocationBox();
    newBox.createLocationBox();
}


class LocationBox {
    constructor(locationName, WPI){
        //this.location = locationName;
        this.location = document.getElementById('cityInput').value;
        //this.WPI      = WPI;
        this.WPI      = document.getElementById('WPIInput').value;
    }

    createLocationBox() {
        const box = document.createElement('div');
        box.className = 'loc-box';
    
        // Use a more semantic element for the location, such as <span> or <div>
        const locationElement = document.createElement('span');
        locationElement.className = 'location';
        locationElement.textContent = this.location;
        box.appendChild(locationElement);
    
        // Create a container for WPI and "/10" to align them correctly
        const wpiContainer = document.createElement('div');
        wpiContainer.className = 'wpi-container';
    
        const wpiElement = document.createElement('span');
        wpiElement.className = 'wpi';
        wpiElement.textContent = this.WPI;
        wpiContainer.appendChild(wpiElement);
    
        const wpiScale = document.createElement('span');
        wpiScale.className = 'wpi-scale';
        wpiScale.textContent = '/10';
        wpiContainer.appendChild(wpiScale);
    
        box.appendChild(wpiContainer);

        // move middle point of grading according to WPI
        let gradingMidpoint = this.WPI * 10;
        box.style.background = `linear-gradient(158deg, #d87d6d, ${gradingMidpoint}%, #012353)`;
        
    
        document.querySelector('.sandboxcontainer').appendChild(box);
    }
}