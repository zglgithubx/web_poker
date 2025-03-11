import React from 'react';
import DraggableCard from './components/DraggableCard';
import useWebSocket from './hooks/useWebSocket';
import './styles/App.css';

const App: React.FC = () => {
    const { sendMessage, messages } = useWebSocket('ws://your-websocket-url');

    const handleCardDrop = (suit: string, rank: string) => {
        sendMessage(JSON.stringify({ suit, rank }));
    };

    return (
        <div className="App">
            <h1>可拖动的扑克牌</h1>
            <div className="card-container">
                <DraggableCard suit="heart" rank="A" onDrop={handleCardDrop} />
                <DraggableCard suit="diamond" rank="K" onDrop={handleCardDrop} />
                <DraggableCard suit="club" rank="10" onDrop={handleCardDrop} />
                <DraggableCard suit="spade" rank="5" onDrop={handleCardDrop} />
                <DraggableCard suit="joker" rank="" onDrop={handleCardDrop} />
            </div>
            {/* <div className="messages">
                {messages.map((msg, index) => (
                    <div key={index}>{msg}</div>
                ))}
            </div> */}
        </div>
    );
};

export default App;