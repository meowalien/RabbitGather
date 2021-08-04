CREATE
  (:User {name: $username})-
  [:Create {timestamp: timestamp()}]
  ->(:Article {
    id:      $id,
    title:   $title,
    content: $content, timestamp: timestamp()
  })
    -[:CreateAt {timestamp: timestamp()}]
    ->(:Position {time: timestamp(), pt: point({longitude: $longitude, latitude: $latitude})});


//CREATE
//  (:User {name: '$username'})-
//  [:Create {timestamp: timestamp()}]
//  ->(:Article {
//    id:      9849851,
//    title:  ' $title',
//    content: '$content', timestamp: timestamp()
//  })
//    -[:CreateAt {timestamp: timestamp()}]
//    ->(:Position {time: timestamp(), pt: point({longitude: 121.51187490970621, latitude: 25.040056717110396})});