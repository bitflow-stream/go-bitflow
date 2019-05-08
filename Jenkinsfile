pipeline {
    agent {
        docker {
            image 'golang:1.12-alpine'
            args '-v /root/.goroot:/go'
        }
    }
    stages {
        stage('Build & test') { 
            steps {
                sh '''
                    go clean -i ./...
                    go install ./...
                    rm -rf reports && mkdir -p reports
                    go test -v ./... 2>&1 | go-junit-report > reports/test.xml
                    go vet ./... &> reports/vet.txt
                    golint $(go list -f '{{.Dir}}' ./...) &> reports/lint.txt
                '''
            }
            post {
                always {
                    archiveArtifacts 'reports/*'

                    // TODO: capture test results. Enable coverage and capture report.
                    // TODO: add static code analysis stage
                }
            }
        }
        stage('Slack message') {
            steps { sh 'true' }
            post {
                success {
                    withSonarQubeEnv('CIT SonarQube') {
                        slackSend color: 'good', message: "Build ${env.JOB_NAME} ${env.BUILD_NUMBER} was successful (<${env.BUILD_URL}|Open Jenkins>) (<${env.SONAR_HOST_URL}|Open SonarQube>)"
                    }
                }
                failure {
                    slackSend color: 'danger', message: "Build ${env.JOB_NAME} ${env.BUILD_NUMBER} failed (<${env.BUILD_URL}|Open Jenkins>)"
                }
            }
        }
    }
}

