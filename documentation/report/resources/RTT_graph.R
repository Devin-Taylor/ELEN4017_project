library(ggplot2)
library(stringr)
library(dplyr)

sampleData <- read.table("../../timer_map.txt")

names(sampleData) <- c("Time", "Protocol", "Connection", "Proxy")
sampleData$Proxy <- as.character(sampleData$Proxy)
sampleData$Proxy[sampleData$Proxy != "off"] <- "on"

sampleData$Time <- as.numeric(str_replace(sampleData$Time, "ms", ""))

sampleData$HTTP_Calls <- paste(sampleData$Protocol, sampleData$Connection, sampleData$Proxy, sep = " ")
sampleData <- dplyr::select(sampleData, HTTP_Calls, Time)

ggplot(data=sampleData, aes(x=HTTP_Calls, y=Time)) +
  geom_bar(stat="identity",color = "black", fill = "#3366FF") + 
  labs(x="Function Call Made (Protocol - Connection type - Proxy)", y = "Time (ms)") + 
  ggtitle("Function calls vs time taken for message sent by client to be received") +
  theme_bw() +
  theme(
        legend.key = element_blank(),
        panel.grid = element_blank(),
        panel.grid.minor = element_blank(), 
        panel.grid.major = element_blank(),
        panel.background = element_blank(),
        plot.background = element_rect(fill = "transparent",colour = NA))

ggsave("RTT.png", bg = "transparent")
