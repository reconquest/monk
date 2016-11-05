# requires setup.sh to be sourced first!

cure_user="cure"

:cure:with-key() {
    :cure \
        -u $cure_user ${ips[*]/#/-o} -k "$(ssh-test:print-key-path)" "${@}"
}

:cure:with-password() {
    local password="$1"
    shift

    :expect() {
        expect -f <(cat) -- "${@}" </dev/tty
    }

    go-test:run :expect -u $cure_user ${ips[*]/#/-o} -p "${@}" <<EXPECT
        spawn -noecho cure {*}\$argv

        expect {
            Password: {
                send "$password\r"
                interact
            } eof {
                send_error "\$expect_out(buffer)"
                exit 1
            }
        }
EXPECT
}

:cure:with-key-passphrase() {
    local passphrase="$1"
    shift

    :expect() {
        expect -f <(cat) -- "${@}" </dev/tty
    }

    go-test:run :expect -u $cure_user ${ips[*]/#/-o} \
            -k "$(ssh-test:print-key-path)" "${@}" <<EXPECT
        spawn -noecho cure {*}\$argv

        expect {
            "Key passphrase:" {
                send "$passphrase\r"
                interact
            } eof {
                send_error "\$expect_out(buffer)"
                exit 1
            }
        }
EXPECT
}

:cure() {
    tests:debug "!!! cure ${@}"

    go-test:run cure "${@}"
}
