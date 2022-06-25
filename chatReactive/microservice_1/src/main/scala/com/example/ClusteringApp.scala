package com.example

import akka.actor.typed.scaladsl.Behaviors
import akka.actor.typed.ActorSystem
import akka.actor.typed.Behavior
import com.typesafe.config.ConfigFactory

object App {

  object RootBehavior {
    def apply(): Behavior[Nothing] = Behaviors.setup[Nothing] { context =>
      // Create an actor that handles cluster domain events
      val clusterListener = context.spawn(ClusterListener(), "ClusterListener")
      val httpServer = context.spawn[Nothing](HttpServerWithActorInteraction(), "HttpServer")

      Behaviors.empty
    }
  }

  def main(args: Array[String]): Unit = {
    
    val config = ConfigFactory.load()
    kamon.Kamon.init(config);

    // Create an Akka system
    ActorSystem[Nothing](RootBehavior(), "clustering-cluster", config)
  }

}
