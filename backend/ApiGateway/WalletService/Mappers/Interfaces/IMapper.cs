namespace WalletService.Mappers.Interfaces
{
    public interface IMapper<E, D>
    {
        E ToEntity(D d);
        D ToDto(E e);
    }
}
