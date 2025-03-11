import React from 'react';
import Card from './Card';

const DraggableCard: React.FC<{ suit: string; rank: string; onDrop: (suit: string, rank: string) => void; }> = ({ suit, rank, onDrop }) => {
    const handleDragStart = (e: React.DragEvent) => {
        e.dataTransfer.setData('text/plain', `${suit}-${rank}`);
    };

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault();
        const data = e.dataTransfer.getData('text/plain');
        const [droppedSuit, droppedRank] = data.split('-');
        onDrop(droppedSuit, droppedRank);
    };

    const handleDragOver = (e: React.DragEvent) => {
        e.preventDefault();
    };

    return (
        <div
            className={`card suit-${suit.toLowerCase()}`}
            draggable
            onDragStart={handleDragStart}
            onDrop={handleDrop}
            onDragOver={handleDragOver}
            style={{ display: 'inline-block', margin: '10px' }}
        >
            <Card suit={suit} rank={rank} />
        </div>
    );
};

export default DraggableCard;