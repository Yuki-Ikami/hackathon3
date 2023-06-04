import React, { useState, useEffect } from 'react';
import axios from 'axios';

interface Channel {
  id: number;
  name: string;
}

interface Message {
  id: number;
  content: string;
}

const Chat: React.FC = () => {
  const [channels, setChannels] = useState<Channel[]>([]);
  const [selectedChannel, setSelectedChannel] = useState<Channel | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [newMessage, setNewMessage] = useState('');

  useEffect(() => {
    // チャンネルの一覧を取得
    axios.get('/api/channels')
      .then(response => {
        setChannels(response.data);
      })
      .catch(error => {
        console.error('Error fetching channels:', error);
      });
  }, []);

  useEffect(() => {
    if (selectedChannel) {
      // 選択されたチャンネルのメッセージを取得
      axios.get(`/api/channels/${selectedChannel.id}/messages`)
        .then(response => {
          setMessages(response.data);
        })
        .catch(error => {
          console.error('Error fetching messages:', error);
        });
    }
  }, [selectedChannel]);

  const handleChannelSelect = (channel: Channel) => {
    setSelectedChannel(channel);
  };

  const handleSendMessage = () => {
    if (newMessage.trim() !== '') {
      // メッセージを送信
      axios.post(`/api/channels/${selectedChannel!.id}/messages`, { content: newMessage })
        .then(response => {
          setNewMessage('');
          // メッセージ一覧を更新
          setMessages(prevMessages => [...prevMessages, response.data]);
        })
        .catch(error => {
          console.error('Error sending message:', error);
        });
    }
  };

  return (
    <div>
      <div>
        <h2>Channels</h2>
        <ul>
          {channels.map(channel => (
            <li key={channel.id} onClick={() => handleChannelSelect(channel)}>
              {channel.name}
            </li>
          ))}
        </ul>
      </div>
      <div>
        <h2>Messages</h2>
        {selectedChannel ? (
          <div>
            <h3>{selectedChannel.name}</h3>
            <ul>
              {messages.map(message => (
                <li key={message.id}>{message.content}</li>
              ))}
            </ul>
            <div>
              <input type="text" value={newMessage} onChange={e => setNewMessage(e.target.value)} />
              <button onClick={handleSendMessage}>Send</button>
            </div>
          </div>
        ) : (
          <p>Select a channel to view messages.</p>
        )}
      </div>
    </div>
  );
};

export default Chat;
