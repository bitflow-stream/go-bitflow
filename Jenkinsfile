pipeline {
    agent {
        docker {
            image 'teambitflow/golang-build:1.12-stretch'
            args '-v /root/.goroot:/go'
        }
    }
    stages {
        stage('Build & test') { 
            steps {
                    sh 'go clean -i -v ./...'
                    sh 'go install -v ./...'
                    sh 'rm -rf reports && mkdir -p reports'
                    sh 'go test -v ./... -coverprofile=reports/test-coverage.txt 2>&1 | go-junit-report > reports/test.xml'
                    sh 'go vet ./... &> reports/vet.txt'
                    sh 'golint $(go list -f "{{.Dir}}" ./...) &> reports/lint.txt'
            }
            post {
                always {
                    archiveArtifacts 'reports/*'
                    junit 'reports/test.xml'
                }
            }
        }
        stage('SonarQube') {
            when {
                branch 'master'
            }
            steps {
                script {
                    // sonar-scanner which don't rely on JVM
                    def scannerHome = tool 'sonar-scanner-linux'
                    withSonarQubeEnv('CIT SonarQube') {
                        sh """
                            ${scannerHome}/bin/sonar-scanner -Dsonar.projectKey=go-bitflow \
                                -Dsonar.sources=. -Dsonar.tests=. \
                                -Dsonar.inclusions="**/*.go" -Dsonar.test.inclusions="**/*_test.go" \
                                -Dsonar.exclusions="**/script/script/internal/**/*.go" \
                                -Dsonar.go.golint.reportPath=reports/lint.txt \
                                -Dsonar.go.govet.reportPaths=reports/vet.txt \
                                -Dsonar.go.coverage.reportPaths=reports/test-coverage.txt \
                                -Dsonar.test.reportPath=reports/test.xml
                        """
                    }
                }
                timeout(time: 30, unit: 'MINUTES') {
                    waitForQualityGate abortPipeline: true
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

