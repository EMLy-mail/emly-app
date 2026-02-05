#define ApplicationName 'EMLy'
#define ApplicationVersion GetVersionNumbersString('EMLy.exe')
#define ApplicationVersion '1.4.1'

[Setup]
AppName={#ApplicationName}
AppVersion={#ApplicationVersion}
DefaultDirName={autopf}\EMLy
OutputBaseFilename={#ApplicationName}_Installer_{#ApplicationVersion}
ArchitecturesInstallIn64BitMode=x64compatible
DisableProgramGroupPage=yes
; Request administrative privileges for HKA to write to HKLM if needed, 
; or use "lowest" if purely per-user, but file associations usually work better with admin rights or proper HKA handling.
PrivilegesRequired=admin
SetupIconFile=..\build\windows\icon.ico
UninstallDisplayIcon={app}\{#ApplicationName}.exe
AppVerName={#ApplicationName} {#ApplicationVersion}

[Files]
; Source path relative to this .iss file (assuming it is in the "installer" folder and build is in "../build")
Source: "..\build\bin\{#ApplicationName}.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\bin\config.ini"; DestDir: "{app}"; Flags: ignoreversion

[Registry]
; 1. Register the .eml extension and point it to our internal ProgID "EMLy.EML"
Root: HKA; Subkey: "Software\Classes\.eml"; ValueType: string; ValueName: ""; ValueData: "{#ApplicationName}.EML"; Flags: uninsdeletevalue
Root: HKA; Subkey: "Software\Classes\.msg"; ValueType: string; ValueName: ""; ValueData: "{#ApplicationName}.MSG"; Flags: uninsdeletevalue

; 2. Define the ProgID with a readable name and icon
Root: HKA; Subkey: "Software\Classes\{#ApplicationName}.EML"; ValueType: string; ValueName: ""; ValueData: "{#ApplicationName} Email Message"; Flags: uninsdeletekey
Root: HKA; Subkey: "Software\Classes\{#ApplicationName}.EML\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#ApplicationName}.exe,0"

Root: HKA; Subkey: "Software\Classes\{#ApplicationName}.MSG"; ValueType: string; ValueName: ""; ValueData: "{#ApplicationName} Outlook Message"; Flags: uninsdeletekey
Root: HKA; Subkey: "Software\Classes\{#ApplicationName}.MSG\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#ApplicationName}.exe,0"

; 3. Define the open command
; "%1" passes the file path to the application
Root: HKA; Subkey: "Software\Classes\{#ApplicationName}.EML\shell\open\command"; ValueType: string; ValueName: ""; ValueData: """{app}\EMLy.exe"" ""%1"""
Root: HKA; Subkey: "Software\Classes\{#ApplicationName}.MSG\shell\open\command"; ValueType: string; ValueName: ""; ValueData: """{app}\EMLy.exe"" ""%1"""

; Optional: Add "Open with EMLy" to context menu explicitly (though file association typically handles the double click)
Root: HKA; Subkey: "Software\Classes\{#ApplicationName}.EML\shell\open"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "{#ApplicationName}"
Root: HKA; Subkey: "Software\Classes\{#ApplicationName}.MSG\shell\open"; ValueType: string; ValueName: "FriendlyAppName"; ValueData: "{#ApplicationName}"

[Icons]
Name: "{autoprograms}\{#ApplicationName}"; Filename: "{app}\{#ApplicationName}.exe"
