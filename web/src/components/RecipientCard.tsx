export default function RecipientCard({
  name,
  wishlist,
}: {
  name: string;
  wishlist: string;
}) {
  return (
    <div className="bg-green-50 border border-green-200 rounded-lg p-4">
      <h2 className="font-bold text-lg mb-2">Твой подопечный: {name}</h2>
      {wishlist && (
        <div>
          <h3 className="font-medium text-sm text-gray-600 mb-1">Вишлист:</h3>
          <p className="text-gray-800 whitespace-pre-wrap">{wishlist}</p>
        </div>
      )}
    </div>
  );
}
