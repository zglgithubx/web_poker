import { defineStore } from 'pinia';
import { WebSocketService } from '../services/websocket';

interface Card {
  value: number;// 牌面值
  suit: string;// 花色
  isFlipped: boolean;// 是否翻开
  rank:  string;// 牌面值
}
interface GameState {
  deck: Card[];//牌
  playerHand: Card[];//玩家牌
  dealerHand: Card[];//庄家牌
  gameStatus: 'betting' | 'playing' | 'dealerTurn' | 'gameOver';//游戏状态
  playerScore: number;//玩家得分
  dealerScore: number;//庄家得分
  currentBet: number;//当前投注
  chips: number;//筹码
  ws: WebSocketService | null;//websocket
}

export const useGameStore = defineStore('game', {
  state: (): GameState => ({
    deck: [],
    playerHand: [],
    dealerHand: [],
    gameStatus: 'betting',
    playerScore: 0,
    dealerScore: 0,
    currentBet: 0,
    chips: 1000,
    ws: null,
  }),

  actions: {
    initializeWebSocket() {
      this.ws = new WebSocketService('ws://localhost:8080/ws');
      this.ws.connect();

      // 设置消息处理器
      this.ws.on('gameState', (data) => {
        this.updateGameState(data);
      });

      this.ws.on('gameStart', (data) => {
        this.gameStatus = 'playing';
        this.updateGameState(data);
      });

      this.ws.on('gameEnd', (data) => {
        this.gameStatus = 'gameOver';
        this.updateGameState(data);
      });
    },

    updateGameState(data: any) {
      if (data.playerHand) this.playerHand = data.playerHand;
      if (data.dealerHand) this.dealerHand = data.dealerHand;
      if (data.chips !== undefined) this.chips = data.chips;
      if (data.currentBet !== undefined) this.currentBet = data.currentBet;
      this.calculateScores();
    },

    placeBet(amount: number) {
      if (amount <= this.chips) {
        this.ws?.send({
          type: 'placeBet',
          data: { amount }
        });
      }
    },

    hit() {
      this.ws?.send({
        type: 'hit',
        data: {}
      });
    },

    stand() {
      this.ws?.send({
        type: 'stand',
        data: {}
      });
    },

    resetGame() {
      this.ws?.send({
        type: 'resetGame',
        data: {}
      });
    },

    // 保留原有的本地计算逻辑
    calculateScores() {
      this.playerScore = this.calculateHandScore(this.playerHand);
      this.dealerScore = this.calculateHandScore(this.dealerHand.filter(card => !card.isFlipped));
    },

    calculateHandScore(hand: Card[]): number {
      let score = 0;
      let aces = 0;

      for (const card of hand) {
        // if (card.rank === 'A') {
        //   aces++;
        // } else if (['K', 'Q', 'J'].includes(card.value)) {
        //   score += 10;
        // } else {
        //   score += parseInt(card.value);
        // }
      }

      for (let i = 0; i < aces; i++) {
        if (score + 11 <= 21) {
          score += 11;
        } else {
          score += 1;
        }
      }

      return score;
    },
  },
}); 