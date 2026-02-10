#define ApplicationName 'EMLy'
#define ApplicationVersion GetVersionNumbersString('EMLy.exe')
#define ApplicationVersion '1.5.4_beta'

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "italian"; MessagesFile: "compiler:Languages\Italian.isl"

[CustomMessages]
; English messages
english.UpgradeDetected=A previous version of {#ApplicationName} (v%1) has been detected.
english.UpgradeMessage=This installer will upgrade your installation to version {#ApplicationVersion}.%n%nYour settings and preferences will be preserved.%n%nDo you want to continue?
english.FreshInstall=Welcome to {#ApplicationName} {#ApplicationVersion} Setup
english.FreshInstallMessage=This will install {#ApplicationName} on your computer.

; Italian messages
italian.UpgradeDetected=È stata rilevata una versione precedente di {#ApplicationName} (v%1).
italian.UpgradeMessage=Questo installer aggiornerà la tua installazione alla versione {#ApplicationVersion}.%n%nLe tue impostazioni e preferenze saranno preservate.%n%nVuoi continuare?
italian.FreshInstall=Benvenuto nell'installazione di {#ApplicationName} {#ApplicationVersion}
italian.FreshInstallMessage=Questo installerà {#ApplicationName} sul tuo computer.

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
WizardStyle=modern dynamic includetitlebar


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
var
  PreviousVersion: String;
  IsUpgrade: Boolean;

// Check if a previous version is installed
function GetPreviousVersion(): String;
var
  RegPath: String;
  Version: String;
begin
  Result := '';
  
  // Check HKLM (system-wide installation)
  RegPath := 'Software\Microsoft\Windows\CurrentVersion\Uninstall\{#ApplicationName}_is1';
  if RegQueryStringValue(HKLM, RegPath, 'DisplayVersion', Version) then
  begin
    Result := Version;
    Exit;
  end;
  
  // Check HKCU (user installation)
  if RegQueryStringValue(HKCU, RegPath, 'DisplayVersion', Version) then
  begin
    Result := Version;
    Exit;
  end;
end;

// Initialize setup and detect upgrade
function InitializeSetup(): Boolean;
var
  Message: String;
begin
  Result := True;
  PreviousVersion := GetPreviousVersion();
  IsUpgrade := (PreviousVersion <> '');
  
  if IsUpgrade then
  begin
    // Show upgrade message
    Message := FmtMessage(CustomMessage('UpgradeDetected'), [PreviousVersion]) + #13#10#13#10 +
               CustomMessage('UpgradeMessage');
    
    if MsgBox(Message, mbInformation, MB_YESNO) = IDNO then
    begin
      Result := False;
    end;
  end;
end;

// Show appropriate welcome message
procedure InitializeWizard();
begin
  if not IsUpgrade then
  begin
    WizardForm.WelcomeLabel2.Caption := CustomMessage('FreshInstallMessage');
  end;
end;

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
