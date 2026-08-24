// Initialize the map
var map = L.map('map').setView([6.616, 3.506], 11);

// Define tile layers
var streetLayer = L.tileLayer('https://api.maptiler.com/maps/streets-v2/{z}/{x}/{y}.png?key=sImSM6zs3ZeSqXT9mhWS', {
    attribution: '<a href="https://www.maptiler.com/copyright/" target="_blank">&copy; MapTiler</a> <a href="https://www.openstreetmap.org/copyright" target="_blank">&copy; OpenStreetMap contributors</a>'
});

var satelliteLayer = L.tileLayer('https://api.maptiler.com/maps/hybrid/{z}/{x}/{y}.jpg?key=sImSM6zs3ZeSqXT9mhWS', {
    attribution: '<a href="https://www.maptiler.com/copyright/" target="_blank">&copy; MapTiler</a> <a href="https://www.openstreetmap.org/copyright" target="_blank">&copy; OpenStreetMap contributors</a>',
    maxZoom: 19
});

// Add default layer
streetLayer.addTo(map);

// View control buttons
document.getElementById('streetView').addEventListener('click', function() {
    map.removeLayer(satelliteLayer);
    map.addLayer(streetLayer);
    this.classList.add('active');
    document.getElementById('satelliteView').classList.remove('active');
});

document.getElementById('satelliteView').addEventListener('click', function() {
    map.removeLayer(streetLayer);
    map.addLayer(satelliteLayer);
    this.classList.add('active');
    document.getElementById('streetView').classList.remove('active');
});

// Marker icon with rotation
var taxiIcon = L.icon({
    iconUrl: 'img/icon.png',
    iconSize: [17, 17],
    iconAnchor: [20, 20],
    className: 'smooth-move smooth-rotate'
});

// Initial marker
var marker = L.marker([6.616, 3.506], {
    icon: taxiIcon,
    rotationAngle: 0,
    rotationOrigin: 'center'
}).addTo(map);

var routingControl = null;
var animationFrameId = null;

// Function to calculate bearing between two points
function getBearing(start, end) {
    const startLat = start.lat * Math.PI / 180;
    const startLng = start.lng * Math.PI / 180;
    const endLat = end.lat * Math.PI / 180;
    const endLng = end.lng * Math.PI / 180;
    
    const y = Math.sin(endLng - startLng) * Math.cos(endLat);
    const x = Math.cos(startLat) * Math.sin(endLat) - 
              Math.sin(startLat) * Math.cos(endLat) * Math.cos(endLng - startLng);
    const bearing = Math.atan2(y, x);
    return (bearing * 180 / Math.PI + 360) % 360;
}

// Function to calculate distance between two points in meters
function getDistance(start, end) {
    const R = 6371000;
    const φ1 = start.lat * Math.PI/180;
    const φ2 = end.lat * Math.PI/180;
    const Δφ = (end.lat - start.lat) * Math.PI/180;
    const Δλ = (end.lng - start.lng) * Math.PI/180;

    const a = Math.sin(Δφ/2) * Math.sin(Δφ/2) +
              Math.cos(φ1) * Math.cos(φ2) *
              Math.sin(Δλ/2) * Math.sin(Δλ/2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));

    return R * c;
}

// Click handler for map
map.on('click', function(e) {
    // Show loading spinner
    document.querySelector('.loading-spinner').style.display = 'block';
    
    // Cancel any existing animation
    if (animationFrameId) {
        cancelAnimationFrame(animationFrameId);
        animationFrameId = null;
    }
    
    // Remove previous routing control if it exists
    if (routingControl) {
        map.removeControl(routingControl);
    }
    
    // Remove previous second marker if it exists
    if (window.secondMarker) {
        map.removeLayer(window.secondMarker);
    }
    
    // Add new destination marker
    window.secondMarker = L.marker([e.latlng.lat, e.latlng.lng]).addTo(map);

    // Create routing control
    routingControl = L.Routing.control({
        waypoints: [
            L.latLng(6.616, 3.506),
            L.latLng(e.latlng.lat, e.latlng.lng)
        ],
        routeWhileDragging: true,
        showAlternatives: false,
        lineOptions: {
            styles: [{color: '#0078FF', opacity: 0.7, weight: 5}]
        }
    }).on('routesfound', function(e) {
        document.querySelector('.loading-spinner').style.display = 'none';
        
        var route = e.routes[0];
        var coordinates = route.coordinates;
        var totalDistance = route.summary.totalDistance;
        var totalTime = Math.max((totalDistance / 11.11) * 1000, 2000);
        
        // Animation variables
        var startTime = performance.now();
        var duration = totalTime;
        var progress = 0;
        
        function animate(currentTime) {
            var elapsed = currentTime - startTime;
            progress = Math.min(elapsed / duration, 1);
            
            // Get current position
            var index = Math.floor(progress * (coordinates.length - 1));
            var currentPos = coordinates[index];
            
            // Update position
            marker.setLatLng(currentPos);
            
            // Update rotation if not first point
            if (index > 0) {
                var bearing = getBearing(coordinates[index-1], currentPos);
                marker.setRotationAngle(bearing);
            }
            
            // Continue animation if not finished
            if (progress < 1) {
                animationFrameId = requestAnimationFrame(animate);
            }
        }
        
        // Start animation
        animationFrameId = requestAnimationFrame(animate);
    }).addTo(map);
});