#compdef goes -p goes-*

# Install this in FPATH/_goes

local _goes_cmd=${words[1]}
local _goes_cur=${words[CURRENT]}
local -a _goes_complete=( $_goes_cmd complete ${words[@]:1} ) 

if [ -z $_goes_cur ] ; then
	_goes_complete+=( '' )
fi

local -a _goes_reply=( $( "${_goes_complete[@]}" ) )

if [ ${#_goes_reply[@]} -ne 0 ] ; then
	compadd -a _goes_reply
fi
