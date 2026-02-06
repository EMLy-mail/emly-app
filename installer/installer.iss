#define ApplicationName 'EMLy'
#define ApplicationVersion GetVersionNumbersString('EMLy.exe')
#define ApplicationVersion '1.4.1'

[Setup]
AppName={#ApplicationName}
AppVersion={#ApplicationVersion}
; Default directory (will be adjusted in code based on installation mode)
; Admin mode: C:\Program Files\EMLy
; User mode:  C:\Users\{username}\AppData\Local\Programs\EMLy
DefaultDirName={autopf}\EMLy
OutputBaseFilename={#ApplicationName}_Installer_{#ApplicationVersion}
ArchitecturesInstallIn64BitMode=x64compatible
DisableProgramGroupPage=yes
; Allow user to choose between admin (system-wide) and user-only install
; "lowest" = does not require admin privileges by default (user mode)
; "dialog" = shows a dialog asking user to choose installation mode
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog
SetupIconFile=..\build\windows\icon.ico
UninstallDisplayIcon={app}\{#ApplicationName}.exe
AppVerName={#ApplicationName} {#ApplicationVersion}

[Files]
; Source path relative to this .iss file (assuming it is in the "installer" folder and build is in "../build")
Source: "..\build\bin\{#ApplicationName}.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\build\bin\config.ini"; DestDir: "{app}"; Flags: ignoreversion

[Registry]
; File associations using HKA (HKEY_AUTO) registry root
; HKA automatically selects the appropriate registry hive:
; - HKLM (HKEY_LOCAL_MACHINE) for admin/system-wide installations
; - HKCU (HKEY_CURRENT_USER) for user-only installations
; This ensures file associations work correctly for both installation modes

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

[Code]
// Override default directory based on installation mode
function GetDefaultDirName(Param: string): string;
begin
  // If installing with admin privileges (system-wide), use Program Files
  if IsAdminInstallMode then
    Result := ExpandConstant('{autopf}\{#ApplicationName}')
  // If installing for current user only, use AppData\Local\Programs
  else
    Result := ExpandConstant('{localappdata}\Programs\{#ApplicationName}');
end;

procedure CurPageChanged(CurPageID: Integer);
begin
  // Update the directory when the directory selection page is shown
  if CurPageID = wpSelectDir then
  begin
    // Only set if user hasn't manually changed it
    if WizardForm.DirEdit.Text = ExpandConstant('{autopf}\{#ApplicationName}') then
      WizardForm.DirEdit.Text := GetDefaultDirName('');
  end;
end;
