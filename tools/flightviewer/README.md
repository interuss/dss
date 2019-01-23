# flightviewer

```flightviewer``` is a lightweight debug-only viewer for information relating to InterUSS Platform public portal operations. It may be considered a very basic and very non-production portal that combines the steps of looking up operators, querying their public_portal endpoints for flights, and retrieving flight_info about each flight (something that would not be done en masse in any production portal).

## Requirements

Details of the requirements to run flightviewer may be found in ```config.py``` but this section provides an overview.

### InterUSS Platform server
flightviewer's first step in providing information about operations is to ask an InterUSS Platform server which USS's are operating in the area of interest. The URL of this InterUSS Platform server must be provided.

### Authorization information
flightviewer interacts with the InterUSS Platform server and individual USS's on behalf of the user (who is representing themself as a public portal provider) with the provided authorization information. A username and password must be provided which will periodically be used to obtain an access token for the InterUSS Platform server and individual USS's from the specified authorization server. The URL of the authorization server must also be provided.

### Server configuration
flightviewer exposes its user interface as a web page. The server and port for this hosting must be specified.

## Running

### Dependencies

The following pip packages must be installed (```pip install <PACKAGE>```):

* ```flask```
* ```iso8601```
* ```jinja2```
* ```pytz```
* ```requests```

### Testing
flightsim may be launched using the Flask webserver by running flightviewer.py directly (```python flightviewer.py <OPTIONS>```).

## Endpoints

When flightviewer is running, the following endpoints will be available at ```http://FLIGHTVIEWER_SERVER:FLIGHTVIEWER_PORT``` (see ```config.py```).

### GET ```/``` or ```/status```
Print simple Ok message with reference to ```/listoperators``` enpoint.

### GET ```/listoperators?center=LATITUDE,LONGITUDE```
Find and report information about all flights in the area near ```center```.

### GET ```/listoperators?area=LAT,LNG,LAT,LNG,...```
Find and report information about all flights in the area bounded by the polygon described in ```area```.
