package completion

import "github.com/spf13/cobra"

var (
	// bashCompletionFlags maps between a flag (ie: namespace) to a custom shell completion
	bashCompletionFlags = map[string]string{
		"namespace": "__kubectl_get_object namespace",
		// Telepresence has one flag (deployment) which works for both k8s Deployment and Openshift DeploymentConfig
		// Hence we combine output for retrieving both object sets
		"deployment": "__kubectl_get_object_combined deployment deploymentconfig",
	}
)

// AddFlagCompletion adds custom bash completion function defined through bashCompletionFlags map.
func AddFlagCompletion(cmd *cobra.Command) {
	for name, completion := range bashCompletionFlags {
		pflag := cmd.Flags().Lookup(name)
		if pflag != nil {
			if pflag.Annotations == nil {
				pflag.Annotations = map[string][]string{}
			}
			pflag.Annotations[cobra.BashCompCustom] = append(
				pflag.Annotations[cobra.BashCompCustom],
				completion,
			)
		}
	}
}

const (
	// BashCompletionFunc the custom bash completion function to complete object.
	// The bash completion mechanism will launch the __custom_func func if it cannot
	// find any completion and pass the object in the ${last_command}
	// variable.
	//
	// Additional tweaks:
	// - namespace-aware completion: whenever namespace has been specified before any namespace-bound parameter (such as --deployment)
	// we are narrowing the results to that specific namespace (see __kubectl_get)
	// - combining object sets: we can combine results of multiple kubectl_ctl get by calling __kubectl_get_object_combined with multiple types combined
	BashCompletionFunc = `
__ike_get_object()
{
	local type=$1
	local template
	template="{{ range .items  }}{{ .metadata.name }} {{ end }}"
	local ike_out
	if ike_out=$(ike ${type} ls -o template --template="${template}" 2>/dev/null); then
		COMPREPLY=( $( compgen -W "${ike_out}" -- "$cur" ) )
	fi
}

__kubectl_get_object_combined()
{
	for type in "$@"
	do
		COMPREPLY+=($(__kubectl_get $type))
	done
}

__kubectl_get_object()
{
    COMPREPLY=($(__kubectl_get $1))
}

__get_selected_namespace() {
	local namespace
	namespace=$(echo "${COMP_WORDS[@]}" | sed -E 's/.*(--namespace[= ]+|-n +)([A-Za-z_-]+).*/\2/')
	if [[ $namespace = *" "* ]]; then
	    namespace=""
	fi
	echo ${namespace}
}

__kubectl_get()
{
	local template="{{ range .items  }}{{ .metadata.name }} {{ end }}"
    local type=$1
    local kubectl_out
	local namespace=$(__get_selected_namespace)
    if kubectl_out=$(kubectl get $( [[ ! -z "${namespace}" ]] && printf %s "-n ${namespace}" ) -o template --template="${template}" ${type} 2>/dev/null); then
        echo $( compgen -W "${kubectl_out}" -- "$cur" )
		return
    fi
	echo ""
}

__custom_func() {
	case ${last_command} in
		*_describe|*_logs)
			obj=${last_command/ike_/};
			obj=${obj/_describe/}; obj=${obj/_logs/};
			__ike_get_object ${obj}
			return
			;;
		*)
			;;
	esac
}
`
	zshInitialization = `
__ike_bash_source() {
	alias shopt=':'
	alias _expand=_bash_expand
	alias _complete=_bash_comp
	emulate -L sh
	setopt kshglob noshglob braceexpand
	source "$@"
}
__ike_type() {
	# -t is not supported by zsh
	if [ "$1" == "-t" ]; then
		shift
		# fake Bash 4 to disable "complete -o nospace". Instead
		# "compopt +-o nospace" is used in the code to toggle trailing
		# spaces. We don't support that, but leave trailing spaces on
		# all the time
		if [ "$1" = "__ike_compopt" ]; then
			echo builtin
			return 0
		fi
	fi
	type "$@"
}
__ike_compgen() {
	local completions w
	completions=( $(compgen "$@") ) || return $?
	# filter by given word as prefix
	while [[ "$1" = -* && "$1" != -- ]]; do
		shift
		shift
	done
	if [[ "$1" == -- ]]; then
		shift
	fi
	for w in "${completions[@]}"; do
		if [[ "${w}" = "$1"* ]]; then
			echo "${w}"
		fi
	done
}
__ike_compopt() {
	true # don't do anything. Not supported by bashcompinit in zsh
}
__ike_ltrim_colon_completions()
{
	if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
		# Remove colon-word prefix from COMPREPLY items
		local colon_word=${1%${1##*:}}
		local i=${#COMPREPLY[*]}
		while [[ $((--i)) -ge 0 ]]; do
			COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
		done
	fi
}
__ike_get_comp_words_by_ref() {
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[${COMP_CWORD}-1]}"
	words=("${COMP_WORDS[@]}")
	cword=("${COMP_CWORD[@]}")
}
__ike_filedir() {
	local RET OLD_IFS w qw
	if [[ "$1" = \~* ]]; then
		# somehow does not work. Maybe, zsh does not call this at all
		eval echo "$1"
		return 0
	fi
	OLD_IFS="$IFS"
	IFS=$'\n'
	if [ "$1" = "-d" ]; then
		shift
		RET=( $(compgen -d) )
	else
		RET=( $(compgen -f) )
	fi
	IFS="$OLD_IFS"
	IFS=","
	for w in ${RET[@]}; do
		if [[ ! "${w}" = "${cur}"* ]]; then
			continue
		fi
		if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
			qw="$(__ike_quote "${w}")"
			if [ -d "${w}" ]; then
				COMPREPLY+=("${qw}/")
			else
				COMPREPLY+=("${qw}")
			fi
		fi
	done
}
__ike_quote() {
	if [[ $1 == \'* || $1 == \"* ]]; then
		# Leave out first character
		printf %q "${1:1}"
	else
	printf %q "$1"
	fi
}
autoload -U +X bashcompinit && bashcompinit
# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
	LWORD='\<'
	RWORD='\>'
fi
__ike_convert_bash_to_zsh() {
	sed \
	-e 's/declare -F/whence -w/' \
	-e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
	-e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
	-e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
	-e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
	-e "s/${LWORD}_filedir${RWORD}/__ike_filedir/g" \
	-e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__ike_get_comp_words_by_ref/g" \
	-e "s/${LWORD}__ltrim_colon_completions${RWORD}/__ike_ltrim_colon_completions/g" \
	-e "s/${LWORD}compgen${RWORD}/__ike_compgen/g" \
	-e "s/${LWORD}compopt${RWORD}/__ike_compopt/g" \
	-e "s/${LWORD}declare${RWORD}/builtin declare/g" \
	-e "s/\\\$(type${RWORD}/\$(__ike_type/g" \
	<<'BASH_COMPLETION_EOF'
`
)
