# cmpe273-assignment1
Virtual Stock Trading System

Execution Instructions:

1. Start Virtual Stock Trading System Server by executing below command:

    go run /path/to/your/folder/VirtualStockTradingServer.go

2. Run client, either to purchase shares or check portfolio
  
   To purchase shares, run below command:

        go run /path/to/your/folder/VirtualStockTradingClient.go "GOOG:45%,YHOO:55%" 55000

        (stocks, percentage and budget are cusomizable in a format like: "Stock1:%, Stock2:%,......" budget)
        
        Output will be like: 
                  Trade ID: 11068
                  Stocks: {[GOOG:39:$626.91 YHOO:985:$30.71]}
                  Uninvested Amount: $301.16406
        
        
    To see portfolio of some trade id, run below command:
    
        go run /path/to/your/folder/VirtualStockTradingClient.go 11068
        
        (customize trade id (11068) for which you want to see the portfolio)
        
        Output will be like:
                  Stocks: &{[GOOG:39:$626.91 YHOO:985:$30.71]}
                  Current Market Value: $54698.836
                  Uninvested Amount: $301.16406
    
    
