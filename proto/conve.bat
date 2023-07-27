@echo off
:: The call source is from nodejs
:: package for https://www.npmjs.com/package/protobufjs
:: liqi.json version v0.10.217.w
:: npm install -g protobufjs-cli
npm install -g protobufjs@7.2.4
npm install -g protobufjs-cli@1.1.1
pbjs -t proto .\liqi.json -o .\liqi.proto