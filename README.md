# Shift4 Local UTG Helper

Shift4 Local UTG Helper meant to be run alongside Shift4's UTG Standalone application. It acts as a CORS server and terminal ID extraction utility. 

The Shift4 UTG software does not have CORS support by default, so the UTG Helper will forword all calls to Shift4's local UTG software and attach CORS headers when responding to your client. 

In addition, when Shift4's UTG software is installed locally with a single PinPad configured per UTG POS terminal, you can call the `GET /terminalId` endpoint to retrieve the currently configured terminalId for your PinPad. This allows your POS client to be unaware of which unique machine it is running on, but still use the correct terminalId connected to that machine.

![postman screenshot](https://user-images.githubusercontent.com/1683528/224466995-88414393-dc02-4899-abfb-b286353a1dbb.png)

# Build

To build the project yourself, run these commands (_note: you only need to run the rsrc command when updating the manifest_)

`rsrc -manifest utg-helper.exe.manifest -ico icon.ico -o utg-helper.syso`

`GOOS=windows GOARCH=amd64 go build -o bin/utg-helper.exe`

# Run

This software can be installed as a Windows Service, or run standalone.

To run standalone, simply launch the executable and choose option 3. You may also choose to immediately start the server by running `utg-helper.exe start`. Similarly, you can skip straight to the `install` or `uninstall` options by including those arguments in place of `start`.

![command prompt on menu 3](https://user-images.githubusercontent.com/1683528/224466930-bdea5d37-146c-4f82-adce-f78be08ab75c.png)

You may change default values by providing flag arguments as seen in the screenshot. It is recommended that you change your originURL to the URL your client will be calling from. For example, if my POS client software is hosted on mywebsite.com, I would run `utg-helper.exe -originURL=mywebsite.com` Note that providing arguments will also allow you to skip the menu and immeditaley start the web server.

![command prompt launched with -help](https://user-images.githubusercontent.com/1683528/224467242-72c93c4b-f464-4d2e-adcb-afb7d48d9bb3.png)

# Install

To install the UTG Helper as a Windows Service, either run the utility with the `install` argument, or select option 1 in the menu. The executable will be copied to `C:\Program Files\Shift4 Helper\utg-helper.exe` and should show up in Windows' Services. After installation, it's recommended that locate it inside Windows' Services (run `services.msc`) and modify the _Startup type_ and _Recovery_ options as you see fit. Note that if you want to append start parameters such as the ones mentioned in the above **Run** section, you must [modify the registry or append them to the executable path](https://serverfault.com/questions/507561/in-a-windows-service-will-the-start-parameters-be-preserved-if-the-start-is-of).

![windows services](https://user-images.githubusercontent.com/1683528/224466483-20691850-2caf-4d51-91eb-b2776a8f9745.png)

# ToDo
-  [ ] Use HTTPS
-  [ ] Optionally disable TLS verification
