pushd serverbrowser
call build.bat
popd
pushd serverstats
call build.bat
popd
rm rs2bundle.zip
zip -j -9 rs2bundle.zip serverbrowser/serverbrowser.exe serverbrowser/flags.csv serverbrowser/flags.png serverbrowser/IP2LOCATION-LITE-DB1.BIN serverbrowser/serverbrowser.json serverstats/serverstats.exe