package com.idugalic

import scala.concurrent.duration._

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import io.gatling.jdbc.Predef._

/**
  * Example Gatling load test that sends HTTP GET requests to a URLs.
  * Run this simulation with:
  * mvn -Dgatling.simulation.name=RecordedSimulation gatling:execute
  *
  * @author Ivan Dugalic
  */
class RecordedSimulation extends Simulation {

    /*
     * A HTTP protocol is used to specify common properties of request(s) to be sent,
     * for instance the base URL, HTTP headers that are to be enclosed with all requests etc.
     */
	val httpProtocol = http
		.baseURL("http://127.0.0.1:8080")
		.inferHtmlResources()
		.acceptHeader("text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		.acceptEncodingHeader("gzip, deflate")
		.acceptLanguageHeader("en-US,en;q=0.5")
		.userAgentHeader("Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:52.0) Gecko/20100101 Firefox/52.0")

	val headers_0 = Map("Upgrade-Insecure-Requests" -> "1")


    /*
     * A scenario consists of one or more requests. For instance logging into a e-commerce
     * website, placing an order and then logging out.
     * One simulation can contain many scenarios.
     */
	val scn = scenario("RecordedSimulation")
		.exec(http("GET - Blog Posts")
			.get("/blogposts")
			.headers(headers_0))
		.pause(0)
		.exec(http("GET - Projects")
			.get("/projects")
			.headers(headers_0))

    /*
     * Define the load simulation.
     * Here we can specify how many users we want to simulate, if the number of users is to increase
     * gradually or if all the simulated users are to start sending requests at once etc.
     * We also specify the HTTP protocol to be used by the load simulation.
     */
	setUp(scn.inject(atOnceUsers(10))).protocols(httpProtocol)
}