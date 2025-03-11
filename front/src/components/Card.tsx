import React from 'react';

interface CardProps {
    suit: string;
    rank: string;
}

const Card: React.FC<CardProps> = ({ suit, rank }) => {
    return (
        <div className={`card suit-${suit.toLowerCase()}`}>
            <div className="top-left">
                <span>{rank}</span>
                <span>{suit}</span>
            </div>
            {/* <div className="center">
                {suit === 'joker' ? '🃏' : suit}
            </div> */}
            <div className="bottom-right">
                <span>{rank}</span>
                <span>{suit}</span>
            </div>
        </div>
    );
};

export default Card;