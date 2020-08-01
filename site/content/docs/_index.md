---
title: 'Overview'
weight: 1
---

AppArea is a simple, easy-to-use tool for casting ports on localhost to a
server accessible by anyone in the outside world. Quite often, when
developing, you'll have a listening server of some sort that you need to show
to a colleague, or a friend, or a client - now you can!

Now, disclaimer, there's a lot of existing software that does this. However,
looking through them, I couldn't find anything that was open source,
self-hosted, and was zero config.

AppArea builds on existing technology, with the server using a lightweight
wrapper above SSH. This means that any existing SSH client (with port
forwarding functionality) can use the server!
