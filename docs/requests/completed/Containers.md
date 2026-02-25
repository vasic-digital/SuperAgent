Add into the execution of all existing test the following entry point test which will be executed
first before any other test! If this test fails we will not continue further with execution of any
next test (or Challenge). It is mandatory (blocker) pre-condition for execution of all types of the
tests we have and execution of the Challenges! First step of the test is to verify existence of .env
file in Containers module (go module, git submodule). If there is none or it is empty, all containers
we have will be booted up on local host! We expect that helixagent when booted will perform startup
of each of them! We will start helixagent (build if needed so it runs ALWAYS the latest codebase) and
verify that it has brought up all containers with the latest codebases versions on current host
(MCPs, LSPs, ACPs, Embeddings, Local RAGs servers included). When helixagent is up with the latest
and greates codebases for itself and all containers test will be executed as completed and all next
tests and Challenges can be executed! If not, test must fail with proper error details so we can
perform diagnostics, research, debugging and final bug fixing! However, if .env file contains
definitions for remote enpoints (see .env.example under Containers), then helixagent when boots will
bring first up all services (all containers) on remote endpoint (host) by distributing it using
already implemented mechnism(s). helixagent will auto-detect all containers being up on the same
network (as we have already implemented and it is supported) and continue working using remote
conatiners running in the same network. We stil do support remote containers running in the cloud
completely outside of our network and already defined rules still apply: first we use for certain
service (container) the ones from the configuration running in the cloud outside of our network (if
they are fully reachable), next priority are discovered ones in same local network, and finally the
lowest priority - services (containers) running on the same host. If we have multiple instances of
the same priority running and being detected and reachable, boot process of helixagent will fail with
proper details so we can investigate, debug and fix! We should not have multiple instances of same
service in the cloud defined in our configuration files, or detected multiple instances of the same
service in the local network or running them in multiple instances on the current host machine! Make
sure we have all this in our system fully implemented, covered with all types of the tests we support
and fully functional. Once all this is implemented, polished and verified, documented up to the
smallest details (documentation fully extended / updated, user guides and manuals, SQL definitions,
diagrams and video courses and other relevant materials we have - website too), run only this first
test so we are sure that whole mechanism is working and all services (containers) have booted as we
are expecting it! As we already do during every helixagent boot sequence, LLMsVerifier will perform
scoring of all acessible models of all available providers. We must extend this to after all this is
done and helixagent running to re-export configurations of all AI CLI Coding Agents discovered on the
system which we support (we support 48+ cli coding agents) and to replace the configuration files
with newly exported and validated and verified configuration files! we have discovered that Crush
JSON configuration file which we are generating is invalid! This MUST BE fixed! More info could be
found on Crush officila GitHub: charmbracelet/crush. All 48+ supported CLI Coding Agents source codes
are located here for the reference if needed and the testing of everything: cli_agents. They are all
Git submodules. Let us know when you do everything and come back with full report on all work done!
Pay attention not to break anything! We cannot leave anything disabled, broken, faulty!
