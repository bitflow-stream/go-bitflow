pipeline {
    agent {
        docker {
            image 'teambitflow/golang-build:1.12-stretch'
            args '-v /root/.goroot:/go -v /var/run/docker.sock:/var/run/docker.sock'
        }
    }
    environment {
        registry = 'teambitflow/go-bitflow'
        registryCredential = 'dockerhub'
        normalImage = '' // Empty variable must be declared here to allow passing an object between the stages.
        staticImage = ''
    }
    stages {
        stage('Git') {
            steps {
                script {
                    env.GIT_COMMITTER_EMAIL = sh(
                        script: "git --no-pager show -s --format='%ae'",
                        returnStdout: true
                        ).trim()
                }
            }
        }
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
                                -Dsonar.exclusions="script/script/internal/*.go" \
                                -Dsonar.go.golint.reportPath=reports/lint.txt \
                                -Dsonar.go.govet.reportPaths=reports/vet.txt \
                                -Dsonar.go.coverage.reportPaths=reports/test-coverage.txt \
                                -Dsonar.test.reportPath=reports/test.xml
                        """
                    }
                }
                timeout(time: 10, unit: 'MINUTES') {
                    waitForQualityGate abortPipeline: true
                }
            }
        }
        stage('Docker build') {
            steps {
                script {
                    normalImage = docker.build registry + ":$BRANCH_NAME-build-$BUILD_NUMBER"
                    staticImage = docker.build registry + ":static-$BRANCH_NAME-build-$BUILD_NUMBER",  '-f static.Dockerfile'
                }
            }
        }
        stage('Docker push') {
            when {
                branch 'master'
            }
            steps {
                script {
                    docker.withRegistry('', registryCredential) {
                        normalImage.push("build-$BUILD_NUMBER")
                        normalImage.push("latest")
                        staticImage.push("static-build-$BUILD_NUMBER")
                        staticImage.push("static")
                    }
                }
            }
        }
    }
    post {
        success {
            withSonarQubeEnv('CIT SonarQube') {
                slackSend channel: '#jenkins-builds-all', color: 'good',
                    message: "Build ${env.JOB_NAME} ${env.BUILD_NUMBER} was successful (<${env.BUILD_URL}|Open Jenkins>) (<${env.SONAR_HOST_URL}|Open SonarQube>)"
            }
        }
        failure {
            slackSend channel: '#jenkins-builds-all', color: 'danger',
                message: "Build ${env.JOB_NAME} ${env.BUILD_NUMBER} failed (<${env.BUILD_URL}|Open Jenkins>)"
        }
        fixed {
            withSonarQubeEnv('CIT SonarQube') {
                slackSend channel: '#jenkins-builds', color: 'good',
                    message: "Thanks to ${env.GIT_COMMITTER_EMAIL}, build ${env.JOB_NAME} ${env.BUILD_NUMBER} was successful (<${env.BUILD_URL}|Open Jenkins>) (<${env.SONAR_HOST_URL}|Open SonarQube>)"
            }
        }
        regression {
            slackSend channel: '#jenkins-builds', color: 'danger',
                message: "What have you done ${env.GIT_COMMITTER_EMAIL}? Build ${env.JOB_NAME} ${env.BUILD_NUMBER} failed (<${env.BUILD_URL}|Open Jenkins>)"
        }
    }
}

