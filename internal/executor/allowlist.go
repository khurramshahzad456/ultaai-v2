package executor

import "strings"

// You can later move this to a config file or database
var allowedCommands = []string{
	"df -h",
	"uptime",
	"free -m",
	"apt install wordpress -y",
	"ls -lah /var/www",
	"systemctl status nginx",
	"service nginx restart",
	"top -bn1 | head -n 10",
	"apt update && apt upgrade -y",
}

// IsCommandAllowed checks if the command is in the whitelist{"output":"\nWARNING: apt does not have a stable CLI interface. Use with caution in scripts.\n\nReading package lists...\nBuilding dependency tree...\nReading state information...\nThe following package was automatically installed and is no longer required:\n  python3-netifaces\nUse 'sudo apt autoremove' to remove it.\nThe following additional packages will be installed:\n  libjs-cropper libjs-prototype libjs-scriptaculous php-getid3 php-mysql\n  php8.3-mysql vorbis-tools wordpress-l10n wordpress-theme-twentytwentythree\nSuggested packages:\n  php-com-dotnet php-rar php-sqlite3 php-curl php-imagick php-ssh2\nThe following NEW packages will be installed:\n  libjs-cropper libjs-prototype libjs-scriptaculous php-getid3 php-mysql\n  php8.3-mysql vorbis-tools wordpress wordpress-l10n\n  wordpress-theme-twentytwentythree\n0 upgraded, 10 newly installed, 0 to remove and 8 not upgraded.\nNeed to get 16.4 MB of archives.\nAfter this operation, 96.3 MB of additional disk space will be used.\nDo you want to continue? [Y/n] Abort.\n","command":"sudo apt install wordpress","explanation":"Installs the WordPress content management systemimran@dell:~$

func IsCommandAllowed(input string) bool {
	input = strings.TrimSpace(input)

	for _, allowed := range allowedCommands {
		if strings.HasPrefix(input, allowed) {
			return true
		}
	}
	return false
}
