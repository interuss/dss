# dizzysim

```dizzysim``` is a simple USS simulator that flies simulated aircraft in circles and exposes an InterUSS Platform public portal interface. It publishes its public portal availability to an InterUSS Platform server and then provides information about flights when queried via its public_portal and flight_info endpoints, which conform to the InterUSS Platform public portal API spec.

## Requirements

Details of the requirements to run dizzysim may be found in ```config.py``` but this section provides an overview.

### InterUSS Platform server
dizzysim must inform other USS's or public portal providers that it has flight operations in a particular area. This is accomplished by informing an InterUSS Platform server of its operations, thus making them accessible to other participants via the grid. The URL of the InterUSS Platform server must be provided.

### Authorization information
dizzysim interacts with the InterUSS Platform server with the provided authorization information. A username and password must be provided which will periodically be used to obtain an access token for the InterUSS Platform server from the specified authorization server. The URL of the authorization server must also be provided.

### Simulated flight configuration
The identity and behavior of simulated aircraft must be defined. This includes a "hanger" of simulated aircraft, plus details about the orbit flights they are to fly.

### Server configuration
In addition to the public_portal and flight_info endpoints necessary to provide public portal information to InterUSS Platform participants, dizzysim also provides simulation control endpoints. Information about the serving machine is necessary for this purpose.

## Running

### Dependencies

The following pip packages must be installed (```pip install <PACKAGE>```):

* ```djangorestframework```
* ```flask```
* ```gunicorn``` (for stable development web server)
* ```pyjwt```
* ```pytz```
* ```requests```

### Testing
dizzysim may be launched in test mode (using the Flask webserver) by running dizzysim.py directly (```python dizzysim.py <OPTIONS>```).

### Development

To launch a stable dizzysim instance for development, follow the instructions below.

* Set all appropriate environment variables as detailed in ```config.py```
* ```gunicorn dizzysim:webapp -b 0.0.0.0:PORT --access-logfile -``` where ```PORT``` is the port number where the dizzysim endpoints should be served

Or, to run in the background while logging to file:

```gunicorn dizzysim:webapp -b 0.0.0.0:8120 --access-logfile - > /path/to/logfile.log 2>&1 &```

Real-time log output can then be viewed with:

```tail -f /path/to/logfile.log```

## Endpoints

When dizzysim is running, the following endpoints will be available at ```DIZZY_BASEURL``` (see ```config.py```). Note that they are all protected by an access token provided by the authorization server, which must be included as a ```Authorization=<TOKEN>``` header or a ```access_token=<TOKEN>``` header.

### GET ```/``` or ```/status```
Print detailed information about all simulated flights currently operating.

### POST ```/launch```
Launch an additional aircraft from the simulated hanger. Until ```/land/<i>``` is called, a new aircraft will be launched shortly after landing of the current aircraft launched because of this command (see ```DIZZY_FLIGHTINTERVAL``` in ```config.py```).

### POST ```/land/<i>```
Land the ith aircraft (i=0 means the first aircraft that was launched via ```/launch```, or the aircraft that was automatically launched later due to this first explicitly launch). No followup flights will automatically be launched due to this landing.

### GET ```/public_portal/<coords>```
Implementation of public_portal endpoint required by InterUSS Platform for public portal participants.

### GET ```/flight_info/<uuid>```
Implementation of flight_info endpoint required by InterUSS Platform for public portal participants.
