library(ggplot2)
library(stringr)
library(dplyr)

sampleData <- read.table("../../timer_map1.txt")

names(sampleData) <- c("Time", "CacheStatus")

sampleData$Time <- as.numeric(str_replace(sampleData$Time, "ms", ""))

ggplot(data=sampleData, aes(x=CacheStatus, y=Time)) +
  geom_bar(stat="identity",color = "black", fill = "#3366FF") + 
  labs(x="Cache status for TCP persistent proxy connection", y = "Time (ms)") + 
  ggtitle("Time difference for cached items") +
  theme_bw() +
  theme(
    legend.key = element_blank(),
    panel.grid = element_blank(),
    panel.grid.minor = element_blank(), 
    panel.grid.major = element_blank(),
    panel.background = element_blank(),
    plot.background = element_rect(fill = "transparent",colour = NA))

ggsave("RTT_cache.png", bg = "transparent")
