#define ApplicationName 'EMLy'
#define ApplicationVersion GetVersionNumbersString('EMLy.exe')
#define ApplicationVersion '1.6.3_beta'

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
; Fixed shared path for all users on the machine.
; The first install must be run as admin (right-click → Run as administrator,
; or deployed via IT tooling) so Windows can create C:\3gIT\EMLy and set ACLs.
; After that, any user can re-run this installer to update without UAC,
; because the [Dirs] section below grants the Users group Modify permission.
DefaultDirName=C:\3gIT\EMLy
OutputBaseFilename={#ApplicationName}_Installer_{#ApplicationVersion}
ArchitecturesInstallIn64BitMode=x64compatible
DisableProgramGroupPage=yes
; No elevation is requested — non-admin users can update once the directory
; and its ACLs have been created by the initial admin install.
PrivilegesRequired=lowest
SetupIconFile=..\build\windows\icon.ico
UninstallDisplayIcon={app}\{#ApplicationName}.exe
AppVerName={#ApplicationName} {#ApplicationVersion}
WizardStyle=modern dynamic includetitlebar


[Dirs]
; Grant the built-in Users group Modify permission on the install directory.
; This is what allows non-admin users to overwrite files during an update
; without needing UAC elevation (the directory already exists from the admin
; first install, and their token has write access via this ACL entry).
Name: "{app}"; Permissions: users-modify

[Files]
; Source path relative to this .iss file (assuming it is in the "installer" folder and build is in "../build")
Source: "..\build\bin\{#ApplicationName}.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\config.ini"; DestDir: "{app}"; Flags: ignoreversion

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
  OldInstallDir: String;       // install path read from the old registry entry
  OldUninstallString: String;  // UninstallString value from the old registry entry
  IsLegacyMigration: Boolean;  // True when old path differs from C:\3gIT\EMLy
  OldInstallInHKLM: Boolean;   // True when the old install was system-wide (admin)

const
  RegKey = 'Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\{#ApplicationName}_is1';
  NewPath = 'C:\\3gIT\\EMLy';

// Strips a trailing backslash for consistent path comparison.
function NormalizePath(const Path: String): String;
begin
  Result := Path;
  if (Length(Result) > 3) and (Result[Length(Result)] = '\') then
    Delete(Result, Length(Result), 1);
end;

// Returns True if Param (e.g. '/FORCEUPGRADE') was passed on the command line.
function CmdLineParamExists(const Param: string): Boolean;
var
  I: Integer;
begin
  Result := False;
  for I := 1 to ParamCount do
    if CompareText(ParamStr(I), Param) = 0 then
    begin
      Result := True;
      Exit;
    end;
end;

// Reads the previous installation details from the registry.
// Checks HKLM first (old admin / Program Files install), then HKCU
// (old user-mode AppData install). Populates OldInstallDir,
// OldUninstallString, OldInstallInHKLM, and IsLegacyMigration.
function GetPreviousVersion(): String;
begin
  Result := '';
  OldInstallDir := '';
  OldUninstallString := '';
  OldInstallInHKLM := False;
  IsLegacyMigration := False;

  if RegQueryStringValue(HKLM, RegKey, 'DisplayVersion', Result) then
  begin
    RegQueryStringValue(HKLM, RegKey, 'InstallLocation', OldInstallDir);
    RegQueryStringValue(HKLM, RegKey, 'UninstallString', OldUninstallString);
    OldInstallInHKLM := True;
  end
  else if RegQueryStringValue(HKCU, RegKey, 'DisplayVersion', Result) then
  begin
    RegQueryStringValue(HKCU, RegKey, 'InstallLocation', OldInstallDir);
    RegQueryStringValue(HKCU, RegKey, 'UninstallString', OldUninstallString);
  end;

  // Flag as a legacy migration when the recorded install directory differs
  // from the new target path so PrepareToInstall knows to remove it first.
  if (OldInstallDir <> '') and
     (CompareText(NormalizePath(OldInstallDir), NormalizePath(NewPath)) <> 0) then
    IsLegacyMigration := True;
end;

function InitializeSetup(): Boolean;
var
  Message: String;
begin
  Result := True;
  PreviousVersion := GetPreviousVersion();
  IsUpgrade := (PreviousVersion <> '');

  // Show the upgrade confirmation dialog only for same-path updates.
  // Legacy migrations are handled silently inside PrepareToInstall.
  if IsUpgrade and not IsLegacyMigration then
  begin
    if not CmdLineParamExists('/FORCEUPGRADE') then
    begin
      Message := FmtMessage(CustomMessage('UpgradeDetected'), [PreviousVersion]) + #13#10#13#10 +
                 CustomMessage('UpgradeMessage');
      if MsgBox(Message, mbInformation, MB_YESNO) = IDNO then
        Result := False;
    end;
  end;
end;

// Called during the "Preparing to Install" phase, before any files are touched.
// Silently removes the legacy installation when it lives at a different path.
function PrepareToInstall(var NeedsRestart: Boolean): String;
var
  UninstExe: String;
  ResultCode: Integer;
begin
  Result := '';

  if not IsLegacyMigration then
    Exit;

  // Removing files from Program Files and deleting HKLM keys both require an
  // elevated process. If we are not elevated, abort with clear instructions
  // rather than letting the old uninstaller fail silently halfway through.
  if OldInstallInHKLM and not IsAdmin() then
  begin
    Result := 'A previous EMLy installation was found in:' + #13#10 +
              '  ' + OldInstallDir + #13#10#13#10 +
              'Removing it requires administrator privileges.' + #13#10 +
              'Please right-click this installer and choose "Run as administrator".';
    Exit;
  end;

  // Run the old InnoSetup uninstaller silently and wait for it to finish
  // before this installer writes any files to the new location.
  // The uninstaller also cleans up the old registry entries (uninstall key,
  // file associations, shortcuts) that were written to HKLM or HKCU.
  UninstExe := RemoveQuotes(OldUninstallString);
  if FileExists(UninstExe) then
  begin
    if not Exec(UninstExe, '/VERYSILENT /SUPPRESSMSGBOXES /NORESTART', '',
                SW_HIDE, ewWaitUntilTerminated, ResultCode) then
    begin
      Result := 'Failed to remove the previous EMLy installation from:' + #13#10 +
                '  ' + OldInstallDir + #13#10#13#10 +
                'Please uninstall it manually, then run this installer again.';
      Exit;
    end;
  end;

  // Remove any files or directories the uninstaller left behind
  // (e.g. config.ini, log files, or the folder itself if non-empty).
  if DirExists(OldInstallDir) then
    DelTree(OldInstallDir, True, True, True);
end;

procedure InitializeWizard();
begin
  // Show the fresh-install welcome text for new installs and migrations alike.
  // For in-place same-path upgrades the default wizard text is appropriate.
  if (not IsUpgrade) or IsLegacyMigration then
    WizardForm.WelcomeLabel2.Caption := CustomMessage('FreshInstallMessage');
end;

