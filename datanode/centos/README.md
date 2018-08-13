# InterUSS Platform evaluation on CentOS
These instructions describe the installation and usage of an InterUSS Platform data node on CentOS without the use of Docker (the preferred method).

## Installation
*These steps must only be followed once on a system.*

1. Install OpenJDK 8:

   ```yum -y install java-1.8.0-openjdk java-1.8.0-openjdk-devel```

1. Do you have Python 2.7 installed?  Check with python --version.  If not, run commands below (many must be as administrative user).

    1. *Building from source (most reliable, least likely to affect other parts of the system):*

       ```yum -y update```  
       ```yum groupinstall -y 'development tools'```  
       ```yum install -y zlib-dev openssl-devel sqlite-devel bzip2-devel wget```  
       ```wget http://www.python.org/ftp/python/2.7.15/Python-2.7.15.tar.xz```  
       ```xz -d Python-2.7.15.tar.xz```  
       ```tar -xvf Python-2.7.15.tar```  
       ```cd Python-2.7.15```  
       ```./configure```  
       ```make && make altinstall```  
       ```/usr/local/bin/python2.7 -m ensurepip --default-pip```  
       ```/usr/local/bin/python2.7 -m pip install --upgrade pip setuptools wheel```  

    1. If you did not follow the instructions in section 2i above, ensure you have pip and wget:

       ```yum install python-pip```  
       ```yum install wget```

1. Create InterUSS Platform Python environment
    1. Change into your preferred home folder (probably ```cd ~```)
    1. Create a virtual environment:

       ```virtualenv interuss_platform```

        1. If you don’t have virtualenv (the above command resulted in an error):
            1. If you installed Python as above:

               ```/usr/local/bin/pip2.7 install virtualenv```

            1. If you already had Python 2 and didn’t follow the steps under part 2:

               ```pip install virtualenv```

                * Note that the above should probably not be run as an administrative user.  If that command doesn’t work, try ```pip install virtualenv --user```
1. Activate the interuss_platform Python environment:

   ```cd interuss_platform```  
   ```source bin/activate```

1. Install Python packages:

   ```pip install --no-cache-dir python-dateutil kazoo flask PyJWT djangorestframework cryptography```

1. Copy InterUSS Platform files

   ```yum install git```  
   ```git clone https://github.com/wing-aviation/InterUSS-Platform```  
   ```cp InterUSS-Platform/datanode/centos/* .```

1. Install Zookeeper:

   ```./install.sh```

1. Edit setup.sh
    1. Open setup.sh in a text editor (perhaps ```vi setup.sh```) and find ```<<REPLACE WITH YOUR PUBLIC KEY>>```
    1. Replace that text with the contents of your OAuth public key.

## Setup
*These steps **must** be followed every time a new shell is started (including system reboots).*

1. Change into interuss_platform folder.  Probably:

    ```cd ~/interuss_platform```

1. Activate Python environment:

   ```source bin/activate```


1. Configure environment variables:

    ```source setup.sh```

## Running
*These steps start up an InterUSS Platform data node, after properly installing and setting up the environment as described in previous sections.*

1. Change into the interuss_platform folder, probably:

   ```cd ~/interuss_platform```

1. Start Zookeeper:

   ```zkServer.sh start```

1. Start storage API per README.md:

   ```python storage_api.py```

## Configuring
### Zookeeper intercommunication
The setup described previously sets up an InterUSS Platform data node whose Zookeeper server is stand-alone -- that is, it does not coordinate with any other data nodes.  Since the heart of the InterUSS Platform is coordination between multiple USSes, a more effective demonstration requires this communication.  To enable InterUSS communication, deploy data nodes as above on at least three servers.  Then, edit zoo.cfg in zoo/zoo_conf.  The last line is 0.0.0.0:2888:3888.  Add additional lines for the other servers directly following this line; e.g., interuss2.arc.nasa.gov:2888:3888 or 192.168.0.78:2888:3888.  The :2888:3888 suffix describes the ports the Zookeeper instances should use to coordinate and should not be changed.  Repeat this step for all servers, bearing in mind that the additional server lines will be different for each server (because which servers are the “other” servers changes according to server).

### Testing without OAuth authentication
If the OAuth token provider is down or not working, allow the storage API to bypass token validation by running with additional arguments: ```python storage_api.py -t sandbox```  In this case, any token with the value “sandbox” will be accepted, or the token may be omitted entirely.

### Stopping
To stop running an InterUSS Platform data node and exit the evaluation environment:

* Ctrl-C or Cmd-C to stop the storage API Python script
* ```zkServer.sh stop``` to stop the Zookeeper instance
* ```deactivate``` to exit the Python virtualenv

### Troubleshooting
Check if the server is operating:

* ```wget http://localhost:8121/status```
  * File named “status” should contain JSON with message=OK, status=success
Check if the Zookeeper server is running:
* ```zkServer.sh status```
  * If “Error contacting service. It is probably not running.” then ```zkServer.sh start-foreground``` (this will display any error messages)
