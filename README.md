# binance trading robot
This repo uses github.com/adshao/go-binance api.

Also here is a blog post https://www.mql5.com/en/blogs/post/747234

> This code is provided as is, without warranty.

> I disclaim responsibility for inappropriate use by users in a real environment.

> The program contains different strategies and indicators that can serve as a template or as a starting point for developers.

Structure:
```
- Expert:
    |- Info Account
    |- Info Exchange
    |- Info Market Watch
    |- Info Symbol
    |- Trade Manager
    |- Strategis: 
        |- Strategy Grid Graph (works)
            |- Strategy Grid (works)
        |- Strategy Ichimoku (unfinished)
        |- Strategy Market Newtral (non real market newtral -just the name-)
        |- Strategy Moving Average (not tested)
        |- Bootstap agregation of the market assets
    |- Indicators.
```

> The expert advisor (as they call it in the metatrader enviroment) works by adding this line to the main program.
```go 
EA := expert.NewExpertAdvisor("apiKey","apiSecret")
```

> If you want to use the application to search for information without starting any strategy, you just have to delete "EA.Init ()" and "EA.Run()". 

>The example "minimum_notional.go" uses the library to obtain the minimum notional of the crosses of the selected currency in terms of any other currency.
