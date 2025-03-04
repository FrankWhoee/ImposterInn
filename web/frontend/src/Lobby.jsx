import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemText from '@mui/material/ListItemText';
import ListItemAvatar from '@mui/material/ListItemAvatar';
import Avatar from '@mui/material/Avatar';

function Player({username, userAvatar}) {
  return (
    <ListItem>
      <ListItemAvatar>
        <Avatar>
          {userAvatar}
        </Avatar>
      </ListItemAvatar>
      <ListItemText primary={username} />
    </ListItem>
  )
}

function Lobby({ ws, setPage, players}) {

  function goBack() {
    console.log("go back")
    setPage("username")
  }

  const playerlist = players.map(player => <Player username={player.username} userAvatar={player.userAvatar} />)

  return (
    <>
      <List sx={{ width: '100%', maxWidth: 360, bgcolor: 'background.paper' }}>
        {playerlist}
      </List>
      <Button variant="contained" onClick={goBack}>
          <ArrowBackIcon />
        </Button>
    </>
  )
}

export default Lobby