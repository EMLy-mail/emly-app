[Setup]
AppName=EMLy
AppVersion=1.2.2
DefaultDirName={autopf}\EMLy
OutputBaseFilename=EMLy_Installer_1.2.2
ArchitecturesInstallIn64BitMode=x64compatible
DisableProgramGroupPage=yes
; Request administrative privileges for HKA to write to HKLM if needed, 
; or use "lowest" if purely per-user, but file associations usually work better with admin rights or proper HKA handling.
PrivilegesRequired=admin
SetupIconFile=..\build\windows\icon.ico
UninstallDisplayIcon={app}\EMLy.exe

[Files]
; Source path relative to this .iss file (assuming it is in the "installer" folder and build is in "../build")
Source: "..\build\bin\EMLy.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\bin\config.ini"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\bin\signed_msg.exe"; DestDir: "{app}"; Flags: ignoreversion

[Registry]
; 1. Register the .eml extension and point it to our internal ProgID "EMLy.EML"
Root: HKA; Subkey: "Software\Classes\.eml"; ValueType: string; ValueName: ""; ValueData: "EMLy.EML"; Flags: uninsdeletevalue
Root: HKA; Subkey: "Software\Classes\.msg"; ValueType: string; ValueName: ""; ValueData: "EMLy.MSG"; Flags: uninsdeletevalue

; 2. Define the ProgID with a readable name and icon
Root: HKA; Subkey: "Software\Classes\EMLy.EML"; ValueType: string; ValueName: ""; ValueData: "EMLy Email Message"; Flags: uninsdeletekey
Root: HKA; Subkey: "Software\Classes\EMLy.EML\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\EMLy.exe,0"

Root: HKA; Subkey: "Software\Classes\EMLy.MSG"; ValueType: string; ValueName: ""; ValueData: "EMLy Outlook Message"; Flags: uninsdeletekey
Root: HKA; Subkey: "Software\Classes\EMLy.MSG\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\EMLy.exe,0"

; 3. Define the open command
; "%1" passes the file path to the application
Root: HKA; Subkey: "Software\Classes\EMLy.EML\shell\open\command"; ValueType: string; ValueName: ""; ValueData: """{app}\EMLy.exe"" ""%1"""
Root: HKA; Subkey: "Software\Classes\EMLy.MSG\shell\open\command"; ValueType: string; ValueName: ""; ValueData: """{app}\EMLy.exe"" ""%1"""

; Optional: Add "Open with EMLy" to context menu explicitly (though file association typically handles the double click)
Root: HKA; Subkey: "Software\Classes\EMLy.EML\shell\open"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "EMLy"
Root: HKA; Subkey: "Software\Classes\EMLy.MSG\shell\open"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "EMLy"

[Icons]
Name: "{autoprograms}\EMLy"; Filename: "{app}\EMLy.exe"
