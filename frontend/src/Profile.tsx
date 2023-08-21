function Profile(obj) {
  const profile = obj.useLoader();

  return (
    <div>
      <h3>{profile.Name}</h3>
    </div>
  );
}

export default Profile;
