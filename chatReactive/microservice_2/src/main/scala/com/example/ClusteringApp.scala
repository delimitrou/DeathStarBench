package com.example

import akka.actor.typed.scaladsl.Behaviors
import akka.actor.typed.ActorSystem
import akka.actor.typed.Behavior
import com.typesafe.config.ConfigFactory

object App {

  object RootBehavior {
    def apply(): Behavior[Nothing] = Behaviors.setup[Nothing] { context =>
      // Create an actor that handles cluster domain events
      val actor = context.spawn(ClusterListener(), "ClusterListener")
      println(actor.path)
      // while(true){
      //   val span = kamon.Kamon.spanBuilder("find-users")
      //     .tag("string-tag", "hello")
      //     .tag("number-tag", 42)
      //     .tag("boolean-tag", true)
      //     .start()

      //   span.finish()
      // }
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
