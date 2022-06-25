/*
 * Copyright (C) 2020-2021 Lightbend Inc. <https://www.lightbend.com>
 */

package com.example

import akka.actor.typed.scaladsl.AskPattern._
import akka.actor.typed.scaladsl.Behaviors
import akka.actor.typed.{ ActorRef, ActorSystem }
import akka.http.scaladsl.Http
import akka.http.scaladsl.marshallers.sprayjson.SprayJsonSupport._
import akka.http.scaladsl.model.StatusCodes
import akka.http.scaladsl.server.Directives._
import akka.util.Timeout
import spray.json.DefaultJsonProtocol._
import akka.actor.typed.Behavior

import scala.concurrent.duration._
import scala.concurrent.{ ExecutionContext, Future }
import scala.io.StdIn

object HttpServerWithActorInteraction {

  object Auction {

    sealed trait Message
    case class Bid(userId: String, offer: Int) extends Message
    case class GetBids(replyTo: ActorRef[Bids]) extends Message
    case class Bids(bids: List[Bid])

    def apply(): Behaviors.Receive[Message] = apply(List.empty)

    def apply(bids: List[Bid]): Behaviors.Receive[Message] = Behaviors.receive {
      case (ctx, bid @ Bid(userId, offer)) =>
        ctx.log.info(s"Bid complete: $userId, $offer")
        apply(bids :+ bid)
      case (_, GetBids(replyTo)) =>
        replyTo ! Bids(bids)
        Behaviors.same
    }

  }

  // these are from spray-json
  implicit val bidFormat = jsonFormat2(Auction.Bid)
  implicit val bidsFormat = jsonFormat1(Auction.Bids)

  def apply(): Behavior[Nothing] = Behaviors.setup[Nothing] { context =>
    val auction: ActorRef[Auction.Message] = context.spawn(Auction(), "auction")
    // needed for the future flatMap/onComplete in the end
    implicit val system = context.system
    implicit val executionContext: ExecutionContext = context.executionContext
    import Auction._

    val route =
      path("auction") {
        concat(
          put {
            parameters("bid") { (bid) =>
              // place a bid, fire-and-forget
              auction ! Bid("Ala", bid.toInt)
              complete(StatusCodes.Accepted, "bid placed")
            }
          },
          get {
            implicit val timeout: Timeout = 5.seconds
            println("get")

            // query the actor for the current auction state
            val bids: Future[Bids] = (auction ? GetBids).mapTo[Bids]
            complete(bids)
          }
        )
      }

    val bindingFuture = Http().newServerAt("0.0.0.0", 8080).bind(route)
    println(s"Server online at http://localhost:8080/")
    // StdIn.readLine() // let it run until user presses return
    // println("Derver end???")
    // bindingFuture
    //   .flatMap(_.unbind()) // trigger unbinding from the port
    //   .onComplete(_ => system.terminate()) // and shutdown when done
    Behaviors.empty
  }
}