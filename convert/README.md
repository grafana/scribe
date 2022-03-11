# Drone to Shipwright converter

This is a prototype of code to convert Drone YAML into Shipwright Golang code.

## Usage
Copy a `.drone.yml` file into this directory, calling it `drone.yml`. Then, in
this directory, run
```
go run .
```
This will output the Golang equivalent of your Drone build.
