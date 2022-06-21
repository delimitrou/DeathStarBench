ThisBuild / scalaVersion := "2.13.8"

val commonScalacSettings = Seq(
  scalacOptions ++= Seq(
    "-deprecation",
    "-unchecked",
    "-encoding", "UTF-8",
    "-Xlint",
  )
)

val akka = "2.6.12"
val telemetry = "1.11.0" // for jaeger tracing

val commonDependencies = Seq (
  // -- Logging --
  "ch.qos.logback" % "logback-classic" % "1.2.3",
  // -- Akka --
  "com.typesafe.akka" %% "akka-actor-typed"   % akka,
  "com.typesafe.akka" %% "akka-cluster-typed" % akka,
  
  // opentelemetry for jaeger
  "io.opentelemetry" % "opentelemetry-bom" % telemetry pomOnly(),
  "io.opentelemetry" % "opentelemetry-api" % telemetry,
  "io.opentelemetry" % "opentelemetry-sdk" % telemetry,
  "io.opentelemetry" % "opentelemetry-exporter-jaeger" % telemetry,
  // "io.opentelemetry" % "opentelemetry-sdk-extension-autoconfigure" % alphaVersion,
  // "io.opentelemetry" % "opentelemetry-exporter-prometheus" % alphaVersion,
  "io.opentelemetry" % "opentelemetry-exporter-zipkin" % telemetry,
  "io.opentelemetry" % "opentelemetry-exporter-jaeger" % telemetry,
  "io.opentelemetry" % "opentelemetry-exporter-otlp" % telemetry,

  "io.opentelemetry.javaagent" % "opentelemetry-javaagent" % telemetry % "runtime"
)

lazy val microservice_1 = project
  .in(file("microservice_1"))
  .enablePlugins(JavaAgent, JavaAppPackaging)
  .settings(
    commonScalacSettings,
    name := "microservice_1",
    libraryDependencies ++= commonDependencies,

    // source & cool example: https://github.com/IvannKurchenko/blog-telemetry
    javaAgents += "io.opentelemetry.javaagent" % "opentelemetry-javaagent" % "1.11.0",
    javaOptions += "-Dotel.javaagent.debug=true",

    version in Docker := "latest",
    dockerExposedPorts in Docker := Seq(1600),
    dockerRepository := Some("suu_project_repository"),
    dockerBaseImage := "java",
  )

lazy val microservice_2 = project
  .in(file("microservice_2"))
  .enablePlugins(JavaAgent, JavaAppPackaging)
  .settings(
    commonScalacSettings,
    name := "microservice_2",
    libraryDependencies ++= commonDependencies,

    // source & cool example: https://github.com/IvannKurchenko/blog-telemetry
    javaAgents += "io.opentelemetry.javaagent" % "opentelemetry-javaagent" % "1.11.0",
    javaOptions += "-Dotel.javaagent.debug=true",

    version in Docker := "latest",
    dockerExposedPorts in Docker := Seq(1600),
    dockerRepository := Some("suu_project_repository"),
    dockerBaseImage := "java"
  )
