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

/* dependencies */
val commonDependencies = Seq (
  // -- Logging --
  "ch.qos.logback" % "logback-classic" % "1.2.3",
  // -- Akka --
  "com.typesafe.akka" %% "akka-actor-typed"   % akka,
  "com.typesafe.akka" %% "akka-cluster-typed" % akka,
)

lazy val microservice_1 = project
  .in(file("microservice_1"))
  .enablePlugins(JavaAppPackaging)
  .settings(
    commonScalacSettings,
    name := "microservice_1",
    libraryDependencies ++= commonDependencies,
    version in Docker := "latest",
    dockerExposedPorts in Docker := Seq(1600),
    dockerRepository := Some("suu_project_repository"),
    dockerBaseImage := "java"
  )

lazy val microservice_2 = project
  .in(file("microservice_2"))
  .enablePlugins(JavaAppPackaging)
  .settings(
    commonScalacSettings,
    name := "microservice_2",
    libraryDependencies ++= commonDependencies,
    version in Docker := "latest",
    dockerExposedPorts in Docker := Seq(1600),
    dockerRepository := Some("suu_project_repository"),
    dockerBaseImage := "java"
  )
