# source this from ~/.bashrc

_goes()
{
	local -a _goes_complete=( ${COMP_WORDS[0]} complete ${COMP_WORDS[@]:1} )
	if [ -z ${COMP_WORDS[COMP_CWORD]} ] ; then
		_goes_complete+=( '' )
	fi
	COMPREPLY=( $(${_goes_complete[@]}) )
	return 0
}

_goes_commands=( goes )
shopt -s nullglob
for d in ${PATH//:/ }; do _goes_commands+=($d/goes-*) ; done
shopt -u nullglob

complete -F _goes -o filenames ${_goes_commands[@]##*/}

unset _goes_commands
