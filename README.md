# Postbear

Postbear is the Postman alternative in your terminal

<img width="1536" height="672" alt="image" src="https://github.com/user-attachments/assets/3b6bffbc-a308-42d4-a4d4-2d90f90037f4" />

---

<img width="1920" height="1050" alt="image" src="https://github.com/user-attachments/assets/dd4a12aa-3c47-4dc2-a7b7-aa48d3679beb" />

## Installation

You can clone this repo and run

```bash
go build -o postbear
```

## Usage
TUI mode
```bash
postbear
```
Read .http file
```bash
postbear read [.http filepath]
```
CLI mode
```bash
postbear run [method] [endpoint]
``` 

## Examples

TUI Mode
![image1](https://github.com/carban/padfadfasboy/blob/main/images/1.gif?raw=true)

TUI Mode reading .http file
![image2](https://github.com/carban/padfadfasboy/blob/main/images/2.gif?raw=true)

CLI Mode
![image3](https://github.com/carban/padfadfasboy/blob/main/images/3.gif?raw=true)

## Command List

| **Command**        	| **Description**                                    	|
|--------------------	|----------------------------------------------------	|
| tab                	| Move Around                                        	|
| shift + tab        	| Reverse Tab                                        	|
| n                  	| New Request (in requests list panel)               	|
| r                  	| Remove Request (in requests list panel)            	|
| enter              	| Send Request                                       	|
| ctrl + s           	| Save Request in a .http file                       	|
| shift + Arrow Keys 	| Change Tabs (Params/Body/Header)                   	|
| enter              	| Move from key input to value input (in Params tab) 	|
| enter              	| Add a new row from value input (in Params tab)     	|
| key up / key down  	| Move around params (in Params tab)                 	|
| ctrl + e           	| Open Environment Variables page                    	|
| ctrl + h           	| Open Help Page                                     	|
| ctrl + c           	| Quit                                               	|

## Acknowledgement

This project were inspired by [Gostman](https://github.com/HalfToothed/gostman)

## Contributing

This is a personal project any feedback is welcome. for major changes, please open an issue first
to discuss what you would like to change.

## License

MIT License
