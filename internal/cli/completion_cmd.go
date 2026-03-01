package cli

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

func completionCmd(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for rctl.

To load completions:

Bash:
  eval "$(rctl completion bash)"

Zsh:
  eval "$(rctl completion zsh)"

Fish:
  rctl completion fish | source
`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(os.Stdout, true)
			case "zsh":
				return writeZshCompletion(os.Stdout)
			case "fish":
				return root.GenFishCompletion(os.Stdout, true)
			default:
				return cmd.Help()
			}
		},
	}
	return cmd
}

// writeZshCompletion writes a custom zsh completion script that displays
// clients/domains (items without descriptions) before subcommands (items with descriptions).
func writeZshCompletion(w io.Writer) error {
	_, err := io.WriteString(w, zshCompletionScript)
	return err
}

const zshCompletionScript = `#compdef rctl
compdef _rctl rctl

__rctl_debug()
{
    local file="$BASH_COMP_DEBUG_FILE"
    if [[ -n ${file} ]]; then
        echo "$*" >> "${file}"
    fi
}

_rctl()
{
    local shellCompDirectiveError=1
    local shellCompDirectiveNoSpace=2
    local shellCompDirectiveNoFileComp=4
    local shellCompDirectiveFilterFileExt=8
    local shellCompDirectiveFilterDirs=16
    local shellCompDirectiveKeepOrder=32

    local lastParam lastChar flagPrefix requestComp out directive comp lastComp noSpace
    local -a completions completions_plain completions_desc

    __rctl_debug "\n========= starting completion logic =========="
    __rctl_debug "CURRENT: ${CURRENT}, words[*]: ${words[*]}"

    words=("${=words[1,CURRENT]}")
    __rctl_debug "Truncated words[*]: ${words[*]},"

    lastParam=${words[-1]}
    lastChar=${lastParam[-1]}
    __rctl_debug "lastParam: ${lastParam}, lastChar: ${lastChar}"

    setopt local_options BASH_REMATCH
    if [[ "${lastParam}" =~ '-.*=' ]]; then
        flagPrefix="-P ${BASH_REMATCH}"
    fi

    requestComp="${words[1]} __complete ${words[2,-1]}"
    if [ "${lastChar}" = "" ]; then
        __rctl_debug "Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi

    __rctl_debug "About to call: eval ${requestComp}"

    out=$(eval ${requestComp} 2>/dev/null)
    __rctl_debug "completion output: ${out}"

    local lastLine
    while IFS='\n' read -r line; do
        lastLine=${line}
    done < <(printf "%s\n" "${out[@]}")
    __rctl_debug "last line: ${lastLine}"

    if [ "${lastLine[1]}" = : ]; then
        directive=${lastLine[2,-1]}
        local suffix
        (( suffix=${#lastLine}+2))
        out=${out[1,-$suffix]}
    else
        __rctl_debug "No directive found.  Setting default"
        directive=0
    fi

    __rctl_debug "directive: ${directive}"
    __rctl_debug "completions: ${out}"

    if [ $((directive & shellCompDirectiveError)) -ne 0 ]; then
        __rctl_debug "Received error directive: aborting."
        return
    fi

    local tab=$'\t'

    while IFS='' read -r comp; do
        if [ -n "$comp" ]; then
            # Convert tab-separated to colon-separated for _describe
            if [[ "$comp" == *"${tab}"* ]]; then
                comp=${comp//"${tab}"/:}
                completions_desc+=("${comp}")
            else
                completions_plain+=("${comp}")
            fi
        fi
    done < <(printf "%s\n" "${out[@]}")

    __rctl_debug "plain completions: ${completions_plain[*]}"
    __rctl_debug "described completions: ${completions_desc[*]}"

    if [ $((directive & shellCompDirectiveNoSpace)) -ne 0 ]; then
        __rctl_debug "Activating nospace."
        noSpace="-S ''"
    fi

    if [ $((directive & shellCompDirectiveFilterFileExt)) -ne 0 ]; then
        local filteringCmd
        filteringCmd='_files'
        completions=("${completions_plain[@]}" "${completions_desc[@]}")
        for filter in ${completions[@]}; do
            if [ ${filter[1]} != '*' ]; then
                filter="\*.$filter"
            fi
            filteringCmd+=" -g $filter"
        done
        filteringCmd+=" ${flagPrefix}"

        __rctl_debug "File filtering command: $filteringCmd"
        _arguments '*:filename:'"$filteringCmd"
    elif [ $((directive & shellCompDirectiveFilterDirs)) -ne 0 ]; then
        completions=("${completions_plain[@]}" "${completions_desc[@]}")
        local subdir
        subdir="${completions[1]}"
        if [ -n "$subdir" ]; then
            __rctl_debug "Listing directories in $subdir"
            pushd "${subdir}" >/dev/null 2>&1
        else
            __rctl_debug "Listing directories in ."
        fi

        local result
        _arguments '*:dirname:_files -/'" ${flagPrefix}"
        result=$?
        if [ -n "$subdir" ]; then
            popd >/dev/null 2>&1
        fi
        return $result
    else
        __rctl_debug "Calling _describe (plain first, then described)"
        local found=0

        # Display plain completions first (clients, domains, commands)
        if [ ${#completions_plain[@]} -gt 0 ]; then
            if eval _describe -V "''" completions_plain $flagPrefix $noSpace; then
                found=1
            fi
        fi

        # Display described completions second (subcommands)
        if [ ${#completions_desc[@]} -gt 0 ]; then
            if eval _describe -V "''" completions_desc $flagPrefix $noSpace; then
                found=1
            fi
        fi

        if [ $found -eq 1 ]; then
            __rctl_debug "_describe found some completions"
            return 0
        else
            __rctl_debug "_describe did not find completions."
            __rctl_debug "Checking if we should do file completion."
            if [ $((directive & shellCompDirectiveNoFileComp)) -ne 0 ]; then
                __rctl_debug "deactivating file completion"
                return 1
            else
                __rctl_debug "Activating file completion"
                _arguments '*:filename:_files'" ${flagPrefix}"
            fi
        fi
    fi
}

if [ "$funcstack[1]" = "_rctl" ]; then
    _rctl
fi
`
