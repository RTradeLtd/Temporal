#!/usr/bin/env bash

# ipfg
# v1.2.2
# IPFS public gateway checker (extended cli version)
# original repository: https://github.com/ipfs/public-gateway-checker
# modifications made to pass shellcheck linting
# this script copyright 2017 by Joss Brown (pseud.): https://github.com/JayBrown
# license: MIT

LANG=en_US.UTF-8
export PATH=/usr/local/bin:$PATH
export IPFS_PATH=/ipfs # This is a modification from Postables @RTrdadeLtd
export HISTIGNORE='*ipfs config show*'
echoerr() {
	echo "$1" 1>&2
}

usage() {
	echo "ipfg $VERSION ($DATE)"
	echo "IPFS public gateway checker (extended cli version)"
	echo ""
	echo "Running 'ipfg' without an option will check the online status of all available IPFS gateways."
	echo ""
	echo "OPTIONS:"
	echo -e "\t-a | --add <URL>\n\t\tAdd specified gateway URLs to local list"
	echo ""
	echo -e "\t-c | --compare\n\t\tCheck remote gateway list for changes incl. auto-backup"
	echo ""
	echo -e "\t-d | --delete [<URL> | all]\n\t\tDelete specified or all gateway URLs from local list"
	echo ""
	echo -e "\t-h | --hash\n\t\tPrint current and saved IPFS hashes incl. auto-backup"
	echo ""
	echo -e "\t-H | --help\n\t\tThis help page"
	echo ""
	echo -e "\t-l | --list [raw] [all | local | remote]\n\t\tDisplay listed gateways as domains or raw URLs"
	echo ""
	echo -e "\t-L | --local [raw]\n\t\tOnly check gateways on local list"
	echo ""
	echo -e "\t-M | --manual <URL>\n\t\tCheck only the specified URLs"
	echo ""
	echo -e "\t-R | --remote [raw]\n\t\tOnly check gateways on remote list"
	echo ""
	echo -e "\t-s | --save [all | hash | url]\n\t\tManually save remote data to local backup files"
	echo ""
	echo -e "\t-u | --upload | --cache [local | remote] <URL>\n\t\tCache an IPFS object from the local or a remote node on a gateway node"
	echo ""
	echo -e "\t-V | --version\n\t\tPrint version number (incl. update check)"
	echo ""
	echo -e "\t-w | --web\n\t\tOpen web-based version of the IPFS public gateway checker"
	echo ""
	echo "Adding 'raw' to a gateway check will print the full URL instead of the gateway domain."
	echo "When using multiple options, only the final option will be recognized."
	echo "Caching will only be attempted on one gateway at a time."
	echo ""
	echo "GATEWAY URL FORMAT:"
	echo -e "\thttp[s]://<domain>.<tld>/ipfs/:hash"
	echo ""
	echo "EXAMPLES:"
	echo -e "\tipfg"
	echo -e "\tipfg -M https://ipfs.io/ipfs/:hash"
	echo -e "\tipfg -l remote"
	echo -e "\tipfg -R raw"
	echo -e "\tipfg -a https://mygateway.com/ipfs/:hash"
	echo -e "\tipfg -l raw local"
	echo -e "\tipfg -L"
	echo -e "\tipfg raw"
	echo -e "\tipfg -d https://mygateway.com/ipfs/:hash"
	echo -e "\tipfg -u local https://ipfs.io/ipfs/\"\$(echo 'This is an ipfg cache run.' | ipfs add -Q)\""
	echo ""
	echo "Copyright 2017 by Joss Brown (pseud.): https://github.com/JayBrown (License: MIT)"
	echo ""
	echo "See also the web-based IPFS gateway checker: https://github.com/ipfs/public-gateway-checker"
}

backup-remote() {
	if ! $INITIAL ; then
		echoerr "Creating backup of remote gateway list. Please wait..."
	fi
	RGATEWAYS=$(curl --silent "$RL_PATH" | sed -e '$ d' -e '1,1d' -e '/^$/d' -e 's/\"//g' -e 's/,$//g')
	if [[ $RGATEWAYS == "" ]] ; then
		echoerr "Error: no access to remote list. Please try again later."
	else
		rm -f "$IPFG_DIR/remote-backup-old" 2>/dev/null \
			&& mv "$RLB_PATH" "$IPFG_DIR/remote-backup-old" 2>/dev/null
		while read -r GATEWAY
		do
			GATEWAY=$(echo "$GATEWAY" | xargs)
			echo "$GATEWAY" >> "$RLB_PATH"
		done < <(echo "$RGATEWAYS")
		echoerr "Done."
	fi
}

backup-hash() {
	if ! $INITIAL ; then
		echoerr "Creating backup of current IPFS hash. Please wait..."
	fi
	WEB_HASH=$(curl --silent "$WB_PATH")
	if [[ $WEB_HASH == "" ]] ; then
		echoerr "Error: no access to current IPFS hash. Please try again later."
		exit 1
	else
		rm -f "$IPFG_DIR/hash-old" 2>/dev/null \
			&& mv "$HASH_PATH" "$IPFG_DIR/hash-old" 2>/dev/null
		echo "$WEB_HASH" > "$HASH_PATH"
		echoerr "Done."
	fi
}

RED='\033[0;31m'
BLUE='\033[0;34m'
GREEN='\033[0;32m'
NC='\033[0m'

VERSION="1.2.2"
DATE="10-2017"

UPDV="1.22"

RL_PATH="https://raw.githubusercontent.com/ipfs/public-gateway-checker/master/gateways.json"
WB_PATH="https://raw.githubusercontent.com/ipfs/public-gateway-checker/master/lastpubver"
REPOV="https://raw.githubusercontent.com/JayBrown/Tools/master/ipfg/ipfg"

TESTHASH="Qmaisz6NMhDB51cCvNWa1GMS7LU1pAxdF4Ld6Ft9kZEP2a"

INITIAL=false

IPFG_DIR="$HOME/.ipfg"
mkdir -p "$IPFG_DIR"

RLB_PATH="$IPFG_DIR/remote-backup"
LL_PATH="$IPFG_DIR/local"
HASH_PATH="$IPFG_DIR/hash"

if ! [[ -f "$LL_PATH" ]] ; then
	touch "$LL_PATH"
fi

if ! [[ -f "$RLB_PATH" ]] ; then
	INITIAL=true
	echoerr "Creating initial backup of remote gateway list. Please wait..."
	backup-remote
	INITIAL=false
fi

if ! [[ -f "$HASH_PATH" ]] ; then
	INITIAL=true
	echoerr "Creating initial backup of current IPFS hash. Please wait..."
	backup-hash
	INITIAL=false
fi

ADD=false
COMPARE=false
DELETE=false
ADDED=false
REMOVED=false
LIST=false
RAW=false
LIST_REMOTE=false
LIST_LOCAL=false
MANUAL=false
REMOTE=false
LOCAL=false
WPRINT=true
LPRINT=true
CACHE=false
CLOCAL=false
CREMOTE=true


if [[ "$*" == "" ]] ; then
	REMOTE=true
	LOCAL=true
else
	if [[ $1 != "-"* ]] ; then
		if [[ $1 == "raw" ]] ; then
			RAW=true
			REMOTE=true
			LOCAL=true
		else
			echoerr "Invalid option: $1"
			echoerr ""
			usage
			exit 1
		fi
	else
		while :; do
			case $1 in
				-a|--add)
					ADD=true
					DELETE=false
					;;
				-c|--compare)
					COMPARE=true
					;;
				-d|--delete)
					DELETE=true
					ADD=false
					;;
				-h|--hash|--hashes)
					WEB_HASH=$(curl --silent "$WB_PATH")
					if [[ $WEB_HASH == "" ]] ; then
						WEB_HASH="n/a"
						WPRINT=false
					fi
					LOCAL_HASH=$(sed -e '/^$/d' "$HASH_PATH")
					if [[ $LOCAL_HASH == "" ]] ; then
						LOCAL_HASH="n/a"
						LPRINT=false
					fi
					if ! $WPRINT && ! $LPRINT ; then
						echoerr "No hashes available. Please try again later."
						exit 1
					else
						echo -e "IPFS hash (current):\t$WEB_HASH"
						echo -e "IPFS hash (backup):\t$LOCAL_HASH"
						if [[ $LOCAL_HASH != "$WEB_HASH" ]] ; then
							echoerr "Attention: a new web version was published."
							backup-hash
						else
							echoerr "No recent changes."
						fi
					fi
					exit 0
					;;
				-H|-\?|--help|--Help)
					usage
					exit 0
					;;
				-l|--list)
					LIST=true
					shift
					if [[ $1 == "raw" ]] ; then
						RAW=true
						shift
					fi
					if [[ $1 == "remote" ]] ; then
						LIST_REMOTE=true
						LIST_LOCAL=false
					elif [[ $1 == "local" ]] ; then
						LIST_REMOTE=false
						LIST_LOCAL=true
					else
						LIST_REMOTE=true
						LIST_LOCAL=true
					fi
					;;
				-L|--local)
					LOCAL=true
					MANUAL=false
					shift
					if [[ $1 == "raw" ]] ; then
						RAW=true
					fi
					;;
				-M|--manual)
					MANUAL=true
					LOCAL=false
					REMOTE=false
					;;
				-R|--remote)
					REMOTE=true
					MANUAL=false
					shift
					if [[ $1 == "raw" ]] ; then
						RAW=true
					fi
					;;
				-s|--save)
					shift
					if [[ $1 == "url" ]] ; then
						backup-remote
					elif [[ $1 == "hash" ]] ; then
						backup-hash
					else
						backup-remote && backup-hash
					fi
					exit 0
					;;
				-u|--upload|--cache)
					CACHE=true
					if [[ $2 == "http"* ]] ; then
						if [[ $2 != *"/ipfs/"* ]] ; then
							echoerr "Error: Not a valid IPFS gateway URL."
							exit 1
						fi
					else
						shift
						if ! [[ $1 =~ ^(local|remote)$ ]] ; then
							echoerr "Error: wrong argument: $1"
							exit 1
						else
							if [[ $1 == "local" ]] ; then
								CLOCAL=true
								# shellcheck disable=SC2034
								CREMOTE=false
							fi
							if [[ $2 != "http"* ]] || [[ $2 != *"/ipfs/"* ]] ; then
								echoerr "Error: Not a valid IPFS gateway URL: $2"
								exit 1
							fi
						fi
					fi
					;;
				-V|--version)
					echo "ipfg $VERSION $DATE"
					echoerr "Checking for updates. Please wait..."
					RVERS=$(curl --silent "$REPOV" | grep "UPDV=" | awk -F\" '{print $2}' | head -1)
					[[ $RVERS == "" ]] && RVERS="1.0"
					if (( $(bc -l <<< "$RVERS > $UPDV") )) ; then
						echo "New update v$RVERS available at: https://github.com/JayBrown/Tools/tree/master/ipfg"
					else
						echoerr "Your version is up-to-date."
					fi
					exit 0
					;;
				-w|--web)
					WEB_HASH=$(curl --silent "$WB_PATH")
					if [[ $WEB_HASH == "" ]] ; then
						WEB_HASH=$(sed -e '/^$/d'"$HASH_PATH")
						if [[ $WEB_HASH == "" ]] ; then
							echoerr "Opening fallback web version of IPFS public gateway checker..."
							open "https://ipfs.github.io/public-gateway-checker/"
							exit 0
						fi
					else
						LOCAL_HASH=$(sed -e '/^$/d' "$HASH_PATH")
						if [[ $LOCAL_HASH != "" ]] && [[ $WEB_HASH != "$LOCAL_HASH" ]] ; then
							echoerr "New hash detected: $WEB_HASH"
							backup-hash
						fi
					fi
					if [[ $(ipfs 2>/dev/null) == "" ]] ; then
						echoerr "Opening web version of IPFS public gateway checker on IPFS gateway..."
						open "https://ipfs.io/ipfs/$WEB_HASH"
					else
						# shellcheck disable=SC2009
						if [[ $(ps aux | grep "ipfs daemon" | grep -v "grep.*ipfs daemon" 2>/dev/null) == "" ]] ; then
							echoerr "Opening web version of IPFS public gateway checker on IPFS gateway..."
							open "https://ipfs.io/ipfs/$WEB_HASH"
						else
							IPFS_PORT=$(ipfs config show | grep "Gateway" | head -1 | rev | awk -F/ '{print $1}' | rev | sed -e 's/,//' -e 's/\"//')
							echoerr "Opening web version of IPFS public gateway checker on localhost..."
							open "http://localhost:$IPFS_PORT/ipfs/$WEB_HASH"
						fi
					fi
					exit 0
					;;
				--)
					shift
					break
					;;
				-?*)
					echoerr "Invalid option: $1"
					echoerr ""
					usage
					exit 1
					;;
				*)
					break
			esac
			shift
		done
	fi
fi

if $ADD ; then
	if [[ "$*" == "" ]] ; then
		echoerr "Error: no gateway URL specified."
		exit 1
	else
		LOCAL_LIST=$(sed -e '/^$/d' "$LL_PATH")
		for ADD_URL in "$@"
		do
			# shellcheck disable=SC2140
			if ! [[ $ADD_URL =~ ^("http://"|"https://") ]] ; then
				echoerr "Error: wrong domain format: $ADD_URL"
				echoerr "Format must be: http[s]://<domain>.<tld>/ipfs/:hash"
				continue
			elif [[ $ADD_URL != *"/ipfs/:hash" ]] ; then
				echoerr "Error: wrong URL format: $ADD_URL"
				echoerr "Format must be: http[s]://<domain>.<tld>/ipfs/:hash"
				continue
			else
				if [[ $(echo "$LOCAL_LIST" | grep ^"$ADD_URL"$) != "" ]] ; then
					echoerr "Already exists: $ADD_URL"
					continue
				else
					echo "$ADD_URL" >> "$LL_PATH"
					echo "Added: $ADD_URL"
				fi
			fi
		done
	fi
	exit 0
fi

if $COMPARE ; then
	RGATEWAYS=$(curl --silent "$RL_PATH" | sed -e '$ d' -e '1,1d' -e '/^$/d' -e 's/\"//g' -e 's/,$//g' | sort)
	if [[ $RGATEWAYS == "" ]] ; then
		echoerr "Error: no access to remote list. Please try again later."
		exit 1
	else
		REMOTELIST=$(while read -r GWURL
			do
				GWURL=$(echo "$GWURL" | xargs)
				echo "$GWURL"
			done < <(echo "$RGATEWAYS")
			)
	fi
	SAVELIST=$(sed -e '/^$/d' "$RLB_PATH" | sort)
	if [[ $SAVELIST == "" ]] ; then
		echoerr "Error: no local backup available."
		exit 1
	fi
	NEWGWS=$(comm -23 <(echo "$REMOTELIST") <(echo "$SAVELIST"))
	if [[ $NEWGWS == "" ]] ; then
		ADDED=false
	else
		ADDED=true
	fi
	DELGWS=$(comm -13 <(echo "$REMOTELIST") <(echo "$SAVELIST"))
	if [[ $DELGWS == "" ]] ; then
		REMOVED=false
	else
		REMOVED=true
	fi
	if ! $ADDED && ! $REMOVED ; then
		echoerr "No changes to remote gateway list."
		exit 0
	else
		echoerr "Changes to remote gateway list:"
		if $ADDED ; then
			while read -r NEWGW
			do
				echo -e "${GREEN}ADDED${NC}\t$NEWGW"
			done < <(echo "$NEWGWS")
		fi
		if $REMOVED ; then
			while read -r DELGW
			do
				echo -e "${RED}REMOVED${NC}\t$DELGW"
			done < <(echo "$DELGWS")
		fi
		backup-remote
	fi
	exit 0
fi

if $DELETE ; then
	LOCAL_LIST=$(sed -e '/^$/d' "$LL_PATH")
	if [[ $LOCAL_LIST == "" ]] ; then
		echoerr "Local gateway list is already empty."
		exit 0
	fi
	# shellcheck disable=SC2199
	if [[ "$@" == "" ]] ; then
		echoerr "Error: no gateway URL specified."
		exit 1
	elif [[ "$@" == "all" ]] ; then
		echo -e "Do you really want to delete all local gateway URL entries? (y/n)"
		read -r -n 1 -s DELYN
		if [[ $DELYN == "y" ]] ; then
			LOCAL_LIST=$(sed -e '/^$/d' "$LL_PATH")
			echo "" > "$LL_PATH"
			while read -r DEL_URL
			do
				echo "Deleted: $DEL_URL"
			done < <(echo "$LOCAL_LIST")
		else
			echoerr "Canceled."
		fi
	else
		for DEL_URL in "$@"
		do
			LOCAL_LIST=$(sed -e '/^$/d' "$LL_PATH")
			# shellcheck disable=SC2140
			if [[ ! $DEL_URL =~ ^("http://"|"https://") ]] ; then
				echoerr "Error: wrong domain format: $DEL_URL"
				echoerr "Format must be: http[s]://<domain>.<tld>/ipfs/:hash"
				continue
			elif [[ $DEL_URL != *"/ipfs/:hash" ]] ; then
				echoerr "Error: wrong URL format: $DEL_URL"
				echoerr "Format must be: http[s]://<domain>.<tld>/ipfs/:hash"
				continue
			elif [[ $(echo "$LOCAL_LIST" | grep "^$DEL_URL$") == "" ]] ; then
				echoerr "Error: not on the local gateway list: $DEL_URL"
				continue
			else
				NEW_LIST=$(echo "$LOCAL_LIST" | grep -v "^$DEL_URL$")
				echo "$NEW_LIST" > "$LL_PATH"
				echo "Deleted: $DEL_URL"
			fi
		done
	fi
	exit 0
fi

if $LIST ; then
	if $LIST_REMOTE ; then
		RGATEWAYS=$(curl --silent "$RL_PATH" | sed -e '$ d' -e '1,1d' -e '/^$/d' -e 's/\"//g' -e 's/,$//g')
		if ! $RAW ; then
			RGATEWAYS=$(echo "$RGATEWAYS" | awk -F/ '{print $1"//"$3}')
		fi
		if [[ $RGATEWAYS == "" ]] ; then
			echoerr "Error: remote gateway list is not accessible."
			echoerr "Checking for backup list..."
			RGATEWAYS=$(awk -F/ '{print $1"//"$3}' "$RLB_PATH")
		fi
		if [[ $RGATEWAYS == "" ]] ; then
			echoerr "Error: no backup list available."
			LIST_REMOTE=false
		else
			RGATEWAYS=$(while read -r GATEWAY
				do
					GATEWAY=$(echo "$GATEWAY" | xargs)
					echo "$GATEWAY"
				done < <(echo "$RGATEWAYS")
				)
		fi
	fi
	if $LIST_LOCAL ; then
		LGATEWAYS=$(sed -e '/^$/d' "$LL_PATH")
		if [[ $LGATEWAYS == "" ]] ; then
			LIST_LOCAL=false
		else
			if ! $RAW ; then
				LGATEWAYS=$(echo "$LGATEWAYS" |  awk -F/ '{print $1"//"$3}')
			fi
		fi
	fi
	if $LIST_REMOTE && $LIST_LOCAL ; then
		echoerr "*** REMOTE ***"
		echo "$RGATEWAYS"
		echoerr "*** LOCAL ***"
		echo "$LGATEWAYS"
	elif $LIST_REMOTE && ! $LIST_LOCAL ; then
		echo "$RGATEWAYS"
	elif ! $LIST_REMOTE && $LIST_LOCAL ; then
		echo "$LGATEWAYS"
	fi
	exit 0
fi

if $CACHE ; then
	if $CLOCAL ; then
		if [[ $(ipfs 2>/dev/null) == "" ]] ; then
			echoerr "Error: ipfs is either not installed or not in your \$PATH"
			exit 1
		fi
		# shellcheck disable=SC2009
		if [[ $(ps aux | grep "ipfs daemon" | grep -v "grep.*ipfs daemon" 2>/dev/null) == "" ]] ; then
			echoerr "Error: IPFS daemon is not running"
			exit 1
		fi
	fi
	CDOMAIN=$(echo "$1" | awk -F/ '{print $1"//"$3}')
	TEMPNAME=$(basename "$1")
	# shellcheck disable=SC2001
	TESTURL=$(echo "$1" | sed "s/$TEMPNAME$/$TESTHASH/")
	if [[ $(curl --silent "$TESTURL") != "Hello from IPFS Gateway Checker" ]] ; then
		RESP_CODE=$(/usr/bin/curl -o /dev/null --silent --head --write-out "%{http_code}\n" "$TESTURL")
		if [[ $RESP_CODE =~ ^(301|303|307|308)$ ]] ; then
			RD_CDOMAIN=$(curl -w "%{url_effective}\n" -I -L -s -S "$CDOMAIN" -o /dev/null | sed 's-/*$--')
			RD_TESTURL="$RD_CDOMAIN/ipfs/$TESTHASH"
			if [[ $(curl --silent "$RD_TESTURL") != "Hello from IPFS Gateway Checker" ]] ; then
				echoerr "Error: gateway on $RD_CDOMAIN [$RESP_CODE] is currently offline."
				exit 1
			else
				CDOMAIN="$RD_CDOMAIN"
				CACHE_URL="$RD_CDOMAIN/ipfs/$TEMPNAME"
			fi
		else
			echoerr "Error: gateway on $CDOMAIN is currently offline [$RESP_CODE]"
			exit 1
		fi
	else
		CACHE_URL="$1"
	fi
	echoerr "Please wait... attempting to cache $TEMPNAME on $CDOMAIN"
	echoerr "Connect timeout: 60 seconds"
	echoerr ""
	curl -o /dev/null --connect-timeout 60 "$CACHE_URL"
	echoerr ""
	echoerr "Done."
	exit 0
fi

if $MANUAL ; then
	# shellcheck disable=SC2199
	if [[ "$@" == "" ]] ; then
		echoerr "Error: no gateway URL specified."
		exit 1
	fi
	GATEWAYS=$(for MANGW in "$@"
		do
			# shellcheck disable=SC2140
			if ! [[ $MANGW =~ ^("http://"|"https://") ]] ; then
				echoerr "Error: wrong domain format: $MANGW"
				echoerr "Format must be: http[s]://<domain>.<tld>/ipfs/:hash"
				continue
			elif [[ $MANGW != *"/ipfs/:hash" ]] ; then
				echoerr "Error: wrong URL format: $MANGW"
				echoerr "Format must be: http[s]://<domain>.<tld>/ipfs/:hash"
				continue
			else
				echo "$MANGW"
			fi
		done
		)
	if [[ $GATEWAYS == "" ]] ; then
		echoerr "Error: no valid gateway URLs specified."
		exit 1
	fi
else
	if $REMOTE ; then
		RGATEWAYS=$(curl --silent "$RL_PATH" | sed -e '$ d' -e '1,1d' -e 's/\"//g' -e 's/,$//g' -e '/^$/d')
		if [[ $RGATEWAYS == "" ]] ; then
			echoerr "Error: remote gateway list is not accessible."
			echoerr "Checking for backup list..."
			RGATEWAYS=$(cat "$RLB_PATH")
		fi
		if [[ $RGATEWAYS == "" ]] ; then
			echoerr "Error: no backup list available."
			REMOTE=false
		fi
	fi

	if $LOCAL ; then
		LGATEWAYS=$(sed -e '/^$/d' "$LL_PATH")
		if [[ $LGATEWAYS == "" ]] ; then
			LOCAL=false
			if ! $REMOTE ; then
				echoerr "No gateway URLs available."
			fi
		fi
	fi

	if ! $REMOTE && ! $LOCAL ; then
		exit 0
	fi

	if $REMOTE && $LOCAL ; then
		GATEWAYS="$RGATEWAYS
$LGATEWAYS"
	elif $REMOTE && ! $LOCAL ; then
		GATEWAYS="$RGATEWAYS"
	elif ! $REMOTE && $LOCAL ; then
		GATEWAYS="$LGATEWAYS"
	fi
fi

GWNUM=$(echo "$GATEWAYS" | wc -l | xargs)
if [[ $GWNUM -gt 1 ]] ; then
	echoerr "*** Public IPFS Gateways ***"
fi

while read -r GATEWAY
do

	GATEWAY=$(echo "$GATEWAY" | xargs)
	# shellcheck disable=SC2001
	TESTURL=$(echo "$GATEWAY" | sed -e "s-:hash-$TESTHASH-")
	if $RAW ; then
		DOMAIN="$TESTURL"
	else
		DOMAIN=$(echo "$GATEWAY" | awk -F/ '{print $1"//"$3}')
	fi

	if [[ $(curl --silent "$TESTURL") == "Hello from IPFS Gateway Checker" ]] ; then
		echo -e "${GREEN}Online${NC}\t$DOMAIN"
	else
		RESP_CODE=$(/usr/bin/curl -o /dev/null --silent --head --write-out "%{http_code}\n" "$TESTURL")
		if [[ $RESP_CODE =~ ^(301|303|307|308)$ ]] ; then
			DOMAIN=$(echo "$GATEWAY" | awk -F/ '{print $1"//"$3}')
			RD_DOMAIN=$(curl -w "%{url_effective}\n" -I -L -s -S "$DOMAIN" -o /dev/null | sed 's-/*$--')
			RD_TESTURL="$RD_DOMAIN/ipfs/$TESTHASH"
			if $RAW ; then
				RD_DOMAIN="$RD_TESTURL"
			fi
			if [[ $(curl --silent "$RD_TESTURL") == "Hello from IPFS Gateway Checker" ]] ; then
				echo -e "${BLUE}Online${NC}\t$RD_DOMAIN\t[$RESP_CODE: $DOMAIN]"
			else
				echo -e "${RED}Offline${NC}\t$RD_DOMAIN\t[$RESP_CODE: $DOMAIN]"
			fi
		else
			echo -e "${RED}Offline${NC}\t$DOMAIN\t[$RESP_CODE]"
		fi
	fi

done < <(echo "$GATEWAYS")

if [[ $GWNUM -gt 1 ]] ; then
	echoerr "Done: $GWNUM gateways checked."
fi
exit 0

