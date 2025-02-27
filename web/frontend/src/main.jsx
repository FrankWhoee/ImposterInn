import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'

var url = "ws://" + window.location.host + "/ws";
var ws = new WebSocket(url);

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <App ws={ws}/>
  </StrictMode>,
)
