ThisBuild / scalaVersion := "2.13.8"

import com.typesafe.sbt.packager.docker._


val commonScalacSettings = Seq(
  scalacOptions ++= Seq(
    "-deprecation",
    "-unchecked",
    "-encoding", "UTF-8",
    "-Xlint",
  )
)

val akka = "2.6.12"
val akkaHttpVersion = "10.2.9"

lazy val commonDependencies = Seq (
  // -- Logging --
  "ch.qos.logback" % "logback-classic" % "1.2.3",
  // -- Akka --
  "com.typesafe.akka" %% "akka-actor-typed"   % akka,
  "com.typesafe.akka" %% "akka-cluster-typed" % akka,
  "com.typesafe.akka" %% "akka-http" % akkaHttpVersion,
  "com.typesafe.akka" %% "akka-http-spray-json" % akkaHttpVersion,
  
  // kamon for jaeger (and potentially prometheus)
  "io.kamon" %% "kamon-bundle" % "2.5.4",
  "io.kamon" %% "kamon-akka" % "2.5.4",
  "io.kamon" %% "kamon-jaeger" % "2.5.4"
)

lazy val microservice_1 = project
  .in(file("microservice_1"))
  .enablePlugins(JavaAppPackaging)
  .settings(
    commonScalacSettings,
    name := "microservice_1",
    libraryDependencies ++= commonDependencies,

    // source & cool example: https://github.com/IvannKurchenko/blog-telemetry
//    javaAgents += "io.opentelemetry.javaagent" % "opentelemetry-javaagent" % "1.11.0",
    javaOptions += "-Dotel.javaagent.debug=true",

    version in Docker := "latest",
    dockerExposedPorts in Docker := Seq(1600),
    dockerRepository := Some("suu_project_repository"),
    dockerBaseImage := "java",

    // run as root 
    daemonUserUid in Docker := Option("0"),
    daemonUser in Docker    := "daemon",

    // dockerCommands ++= Seq(
    //     Cmd("RUN", "wget https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/latest/download/opentelemetry-javaagent.jar"),
    //     Cmd("ENV", "JAVA_OPTS=\"-javaagent:/opt/docker/opentelemetry-javaagent.jar\"")
    // )
  )

lazy val microservice_2 = project
  .in(file("microservice_2"))
  .enablePlugins(JavaAppPackaging)
  .settings(
    commonScalacSettings,
    name := "microservice_2",
    libraryDependencies ++= commonDependencies,

    // source & cool example: https://github.com/IvannKurchenko/blog-telemetry
//    javaAgents += "io.opentelemetry.javaagent" % "opentelemetry-javaagent" % "1.11.0",
    javaOptions += "-Dotel.javaagent.debug=true",

    version in Docker := "latest",
    dockerExposedPorts in Docker := Seq(1601),
    dockerRepository := Some("suu_project_repository"),
    dockerBaseImage := "java",

    // run as root 
    daemonUserUid in Docker := Option("0"),
    daemonUser in Docker    := "daemon",

    // dockerCommands ++= Seq(
    //     Cmd("RUN", "wget https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/latest/download/opentelemetry-javaagent.jar"),
    //     Cmd("ENV", "JAVA_OPTS=\"-javaagent:/opt/docker/opentelemetry-javaagent.jar -Dotel.javaagent.debug=true\""),
    //     // Cmd("ENV", "otel.javaagent.debug=true")
    // )
  )
