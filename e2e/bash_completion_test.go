package e2e_test

import (
	. "github.com/maistra/istio-workspace/e2e/infra"
	"github.com/maistra/istio-workspace/test"
	"github.com/maistra/istio-workspace/test/shell"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bash Completion Tests", func() {

	Context("basic completion", func() {

		It("should show all visible main commands", func() {
			completionResults := completionFor("ike ")
			Expect(completionResults).To(ConsistOf("completion", "develop", "install-operator", "serve", "version", "create", "delete"))
		})

		Context("develop", func() {

			It("should show only required flags for plain command", func() {
				Expect(completionFor("ike develop ")).To(ConsistOf("--deployment=", "-d", "-r", "--run="))
			})

			It("should show all flags only after required ones are passed", func() {
				completionResults := completionFor("ike develop -d deployment -r run.sh -")
				Expect(completionResults).To(ContainElement("--build"))
				Expect(completionResults).To(ContainElement("--route"))
				Expect(completionResults).To(ContainElement("-p"))
				Expect(completionResults).To(ContainElement("--watch"))
			})
		})

		Context("install-operator", func() {

			It("should show matching command", func() {
				Expect(completionFor("ike install-")).To(ConsistOf("install-operator"))
			})

			It("should show all flags inherited from root", func() {
				Expect(completionFor("ike install-operator -")).To(ConsistOf(
					"--config", "--config=", "-c", "--local", "-l", "--namespace", "--namespace=", "-n"))
			})
		})

	})

	// see setup in e2e_suite_test.go#createProjectsForCompletionTests
	Context("kubectl related completion", func() {

		It("should show available namespaces", func() {
			nsCompletion := completionFor("ike develop -n ")
			Expect(nsCompletion).To(ContainElement(CompletionProject1))
			Expect(nsCompletion).To(ContainElement(CompletionProject2))
		})

		It("should show available deployments for current namespace (datawire-project)", func() {
			<-shell.Execute("oc project " + CompletionProject1).Done()
			Expect(completionFor("ike develop -d ")).To(ConsistOf("my-datawire-deployment"))
		})

		It("should show available deployments for selected namespace (datawire-other-project)", func() {
			Expect(completionFor("ike develop -n " + CompletionProject2 + " -d ")).To(ConsistOf("other-1-datawire-deployment", "other-2-datawire-deployment"))
		})
	})

})

func completionFor(cmd string) []string {
	tmpDir := test.TmpDir(GinkgoT(), "ike-bash-completion")
	completionScript := tmpDir + "/get_completion.sh"
	CreateFile(completionScript, getCompletionBash)

	defer DeleteFile(completionScript)

	completion := shell.ExecuteInDir(".", "bash", "-c", ". <(ike completion bash) && source "+completionScript+" && get_completions ' "+cmd+"'")
	<-completion.Done()

	return completion.Status().Stdout
}

const getCompletionBash = `
#
# Author: Brian Beffa <brbsix@gmail.com>
# Original source: https://brbsix.github.io/2015/11/29/accessing-tab-completion-programmatically-in-bash/
# License: LGPLv3 (http://www.gnu.org/licenses/lgpl-3.0.txt)
#

get_completions(){
    local completion COMP_CWORD COMP_LINE COMP_POINT COMP_WORDS COMPREPLY=()

    # load bash-completion if necessary
    declare -F _completion_loader &>/dev/null || {
        source /usr/share/bash-completion/bash_completion
    }

    COMP_LINE=$*
    COMP_POINT=${#COMP_LINE}

    eval set -- "$@"

    COMP_WORDS=("$@")

    # add '' to COMP_WORDS if the last character of the command line is a space
    [[ ${COMP_LINE[@]: -1} = ' ' ]] && COMP_WORDS+=('')

    # index of the last word
    COMP_CWORD=$(( ${#COMP_WORDS[@]} - 1 ))

    # determine completion function
    completion=$(complete -p "$1" 2>/dev/null | awk '{print $(NF-1)}')

    # run _completion_loader only if necessary
    [[ -n $completion ]] || {

        # load completion
        _completion_loader "$1"

        # detect completion
        completion=$(complete -p "$1" 2>/dev/null | awk '{print $(NF-1)}')

    }

    # ensure completion was detected
    [[ -n $completion ]] || return 1

    # execute completion function
    "$completion"

    # print completions to stdout
    printf '%s\n' "${COMPREPLY[@]}" | LC_ALL=C sort
}
`
