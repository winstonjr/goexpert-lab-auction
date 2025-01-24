# Laboratório: Concorrência com Golang - Leilão

### Execução do projeto
Para executar o projeto só é necessário rodar o seguinte comando no diretório raiz da applicação:
```shell
docker compose up
```

### Testes da API
Dentro da pasta `api` existe o arquivo `auctions.http`. Lá existem 3 chamadas http:
- Criação de Leilão
- Buscar Leilões Ativos
- Buscar Leilões Inativos

### Arquivo `.env`
A configuração `AUCTION_INTERVAL` que dita o tempo do leilão está configurada como 1 minuto.
A configuração `CHECK_EXPIRED_AUCTIONS` que dita de quanto em quanto tempo serão verificados os leilões inativos está configurada como 1 minuto.