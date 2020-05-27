#!groovy

import groovy.transform.Field

@Field String email_to = 'sw@platinasystems.com'
@Field String email_from = 'jenkins-bot@platinasystems.com'
@Field String email_reply_to = 'no-reply@platinasystems.com'

pipeline {
    agent any
    stages {
	stage('Build') {
	    steps {
		echo "Building goes..."
		sh 'set +x; for package in `find . -type d -print` ; do ls $package/*.go > /dev/null 2>&1 && { echo "Working on" $package ; { ls $package/*_test.go > /dev/null 2>&1 && { go test -v $package || exit; } } ; { go build  -v -buildmode=archive $package || exit; } } || echo "Skipping" $package;done'
	    }
	}
    }

    post {
	success {
	    mail body: "GOES build ok: ${env.BUILD_URL}",
		from: email_from,
		replyTo: email_reply_to,
		subject: 'GOES build ok',
		to: email_to
	}
	failure {
	    cleanWs()
	    mail body: "GOES build error: ${env.BUILD_URL}",
		from: email_from,
		replyTo: email_reply_to,
		subject: 'GOES BUILD FAILED',
		to: email_to
	}
    }
}
